package logic

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/client/event"
	"github.com/LagrangeDev/LagrangeGo/message"
	"llma.dev/utils/llog"
)

// MessageContext 消息上下文
type MessageContext struct {
	Client   *client.QQClient
	Message  any
	Metadata map[string]any
	ctx      context.Context
}

// NewMessageContext 创建新的消息上下文
func NewMessageContext(client *client.QQClient, msg any) *MessageContext {
	return &MessageContext{
		Client:   client,
		Message:  msg,
		Metadata: make(map[string]any),
		ctx:      context.Background(),
	}
}

// GetContext 获取上下文
func (mc *MessageContext) GetContext() context.Context {
	return mc.ctx
}

// WithContext 设置上下文
func (mc *MessageContext) WithContext(ctx context.Context) *MessageContext {
	mc.ctx = ctx
	return mc
}

// Set 设置元数据
func (mc *MessageContext) Set(key string, value any) {
	mc.Metadata[key] = value
}

// Get 获取元数据
func (mc *MessageContext) Get(key string) (any, bool) {
	value, exists := mc.Metadata[key]
	return value, exists
}

// GetString 获取字符串类型元数据
func (mc *MessageContext) GetString(key string) string {
	if value, exists := mc.Metadata[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetPrivateMessage 获取私聊消息
func (mc *MessageContext) GetPrivateMessage() (*message.PrivateMessage, bool) {
	if msg, ok := mc.Message.(*message.PrivateMessage); ok {
		return msg, true
	}
	return nil, false
}

// GetGroupMessage 获取群消息
func (mc *MessageContext) GetGroupMessage() (*message.GroupMessage, bool) {
	if msg, ok := mc.Message.(*message.GroupMessage); ok {
		return msg, true
	}
	return nil, false
}

// GetFriendRequest 获取好友请求
func (mc *MessageContext) GetFriendRequest() (*event.NewFriendRequest, bool) {
	if msg, ok := mc.Message.(*event.NewFriendRequest); ok {
		return msg, true
	}
	return nil, false
}

// GetMessageText 获取消息文本内容
func (mc *MessageContext) GetMessageText() string {
	if privateMsg, ok := mc.GetPrivateMessage(); ok {
		return extractTextFromElements(privateMsg.Elements)
	}
	if groupMsg, ok := mc.GetGroupMessage(); ok {
		return extractTextFromElements(groupMsg.Elements)
	}
	return ""
}

// reply 回复消息
func (mc *MessageContext) Reply(elements []message.IMessageElement) {
	if privateMsg, ok := mc.GetPrivateMessage(); ok {
		mc.Client.SendPrivateMessage(privateMsg.Sender.Uin, elements)
	}
	if groupMsg, ok := mc.GetGroupMessage(); ok {
		mc.Client.SendGroupMessage(groupMsg.GroupUin, elements)
	}
}

// extractTextFromElements 从消息元素中提取文本
func extractTextFromElements(elements []message.IMessageElement) string {
	var textParts []string
	for _, element := range elements {
		if textElement, ok := element.(*message.TextElement); ok {
			textParts = append(textParts, textElement.Content)
		}
	}
	return strings.Join(textParts, "")
}

// 等待用户确认
func (ctx *MessageContext) Prompt(actionName string, timeout time.Duration, confirmFunc func(), cancelFunc func()) {
	if actionName == "" {
		actionName = "未知操作"
	}
	ctx.Reply([]message.IMessageElement{message.NewText(fmt.Sprintf("你正在执行 %s 请在 %s 内发送“确认”以执行操作", actionName, timeout.String()))})

	// 计算期望的会话标识
	var sm *SessionMatcher

	if pm, ok := ctx.GetPrivateMessage(); ok {
		sm = NewSessionMatcher(PrivateMsg, pm.Sender.Uin, pm.Sender.Uin)
	}
	if gm, ok := ctx.GetGroupMessage(); ok {
		sm = NewSessionMatcher(GroupMsg, gm.GroupUin, gm.Sender.Uin)
	}

	decisionChan := make(chan string, 1)
	// 事件处理器：仅同会话消息，且文本匹配确认/取消时生效
	handler := func(_ context.Context, event Event) error {
		msgEvent, ok := event.(*MessageEvent)
		if !ok {
			llog.Errorf("[router.prompt] Event解析错误")
			return nil
		}
		mc := msgEvent.MessageContext
		if !sm.Match(mc) {
			return nil
		}

		text := strings.TrimSpace(strings.ToLower(mc.GetMessageText()))
		switch text {
		case "确认":
			select {
			case decisionChan <- "confirm":
			default:
			}
		default:
			select {
			case decisionChan <- "cancel":
			default:
			}
		}
		return nil
	}

	GlobalEventBus.Subscribe(EventTypeMessageReceived, handler)

	var decision string
	select {
	case decision = <-decisionChan:
	case <-time.After(timeout):
		decision = "timeout"
	}

	GlobalEventBus.Unsubscribe(EventTypeMessageReceived, handler)

	switch decision {
	case "confirm":
		confirmFunc()
	case "cancel":
		cancelFunc()
	case "timeout":
		ctx.Reply([]message.IMessageElement{message.NewText(fmt.Sprintf("等待超时，已取消 %s 操作", actionName))})
	}
}

// HandlerFunc 处理器函数类型
type HandlerFunc func(ctx *MessageContext) error

// Middleware 中间件类型
type Middleware func(HandlerFunc) HandlerFunc

// Handler 处理器接口
type Handler interface {
	Handle(ctx *MessageContext) error
}

// HandlerAdapter 处理器适配器
type HandlerAdapter struct {
	handler HandlerFunc
}

// NewHandlerAdapter 创建处理器适配器
func NewHandlerAdapter(handler HandlerFunc) *HandlerAdapter {
	return &HandlerAdapter{handler: handler}
}

// Handle 实现Handler接口
func (ha *HandlerAdapter) Handle(ctx *MessageContext) error {
	return ha.handler(ctx)
}

// Route 路由结构
type Route struct {
	Name        string
	Pattern     string
	Handler     Handler
	Middlewares []Middleware
	Matchers    []Matcher
}

// NewRoute 创建新路由
func NewRoute(name string, handler Handler) *Route {
	return &Route{
		Name:        name,
		Handler:     handler,
		Middlewares: make([]Middleware, 0),
		Matchers:    make([]Matcher, 0),
	}
}

// Use 添加中间件
func (r *Route) Use(middleware Middleware) *Route {
	r.Middlewares = append(r.Middlewares, middleware)
	return r
}

// Match 添加匹配器
func (r *Route) Match(matcher Matcher) *Route {
	r.Matchers = append(r.Matchers, matcher)
	return r
}

// SetPattern 设置模式
func (r *Route) SetPattern(pattern string) *Route {
	r.Pattern = pattern
	return r
}

// Execute 执行路由
func (r *Route) Execute(ctx *MessageContext) error {
	// 检查所有匹配器
	for _, matcher := range r.Matchers {
		if !matcher.Match(ctx) {
			return nil // 不匹配，跳过
		}
	}

	// 构建中间件链
	handler := r.Handler.Handle
	for i := len(r.Middlewares) - 1; i >= 0; i-- {
		handler = r.Middlewares[i](handler)
	}

	return handler(ctx)
}

// Router 路由器
type Router struct {
	routes       []*Route
	middlewares  []Middleware
	errorHandler func(error, *MessageContext)
	mu           sync.RWMutex
}

// NewRouter 创建新路由器
func NewRouter() *Router {
	return &Router{
		routes:      make([]*Route, 0),
		middlewares: make([]Middleware, 0),
		errorHandler: func(err error, ctx *MessageContext) {
			llog.Errorf("[lagrange.路由] 处理消息时发生错误: %v", err)
		},
	}
}

// Use 添加全局中间件
func (router *Router) Use(middleware Middleware) *Router {
	router.mu.Lock()
	defer router.mu.Unlock()
	router.middlewares = append(router.middlewares, middleware)
	return router
}

// AddRoute 添加路由
func (router *Router) AddRoute(route *Route) *Router {
	router.mu.Lock()
	defer router.mu.Unlock()
	router.routes = append(router.routes, route)
	return router
}

// Handle 处理消息
func (router *Router) Handle(ctx *MessageContext) {
	router.mu.RLock()
	routes := make([]*Route, len(router.routes))
	copy(routes, router.routes)
	middlewares := make([]Middleware, len(router.middlewares))
	copy(middlewares, router.middlewares)
	router.mu.RUnlock()

	// 为每个路由执行处理
	for _, route := range routes {
		// 创建完整的中间件链（全局中间件 + 路由中间件）
		handler := route.Handler.Handle

		// 先添加路由中间件
		for i := len(route.Middlewares) - 1; i >= 0; i-- {
			handler = route.Middlewares[i](handler)
		}

		// 再添加全局中间件
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}

		// 检查路由匹配
		matched := true
		for _, matcher := range route.Matchers {
			if !matcher.Match(ctx) {
				matched = false
				break
			}
		}

		if matched {
			if err := handler(ctx); err != nil {
				router.errorHandler(err, ctx)
			}
		}
	}
}

// SetErrorHandler 设置错误处理器
func (router *Router) SetErrorHandler(handler func(error, *MessageContext)) {
	router.errorHandler = handler
}

// GetRoutes 获取所有路由
func (router *Router) GetRoutes() []*Route {
	router.mu.RLock()
	defer router.mu.RUnlock()
	routes := make([]*Route, len(router.routes))
	copy(routes, router.routes)
	return routes
}
