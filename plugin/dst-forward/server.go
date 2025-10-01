package dstforward

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/LagrangeDev/LagrangeGo/message"
	"github.com/gin-gonic/gin"
	"llma.dev/bot"
	"llma.dev/config"
	"llma.dev/utils/llog"
)

type MsgType int

const (
	MsgText MsgType = 0 // 消息
	MsgCmd  MsgType = 1 // 命令
)

type MsgQueue struct {
	Messages []Message
	MaxSize  int
}

// Message 消息结构
type Message struct {
	// Type 0 消息 1 命令
	Type MsgType `json:"type"`
	// Data 主要数据
	Data Data `json:"data"`
}

// Data 消息主要数据
type Data struct {
	// Source 来源信息，群组号
	Source Source `json:"source"`
	// Sender 发送者信息
	Sender Sender `json:"sender"`
	// Head 消息是命令类型时会存在head，代表命令的种类，如 save
	Head string `json:"head"`
	// Content 消息正文
	Content any `json:"content"`
}

// Source 来源信息
type Source struct {
	// ID 群组号
	ID uint32 `json:"id"`
	// Name 群组名称
	Name string `json:"name"`
}

// Sender 发送者信息
type Sender struct {
	// ID QQ号码
	ID uint32 `json:"id"`
	// Name QQ用户名
	Name string `json:"name"`
	// Nick 群昵称 (可选)
	Nick string `json:"nick,omitempty"`
}

func DefaultQueue() MsgQueue {
	return MsgQueue{MaxSize: 5}
}

// 入队
func (m *MsgQueue) enqueue(msg Message) {
	llog.Debugf("[dst forward队列] 插入队列消息: %v", msg)
	m.Messages = append(m.Messages, msg)
	if len(m.Messages) > m.MaxSize {
		m.Messages = m.Messages[len(m.Messages)-m.MaxSize : len(m.Messages)]
	}
}

// 全取出
func (m *MsgQueue) drain() []Message {
	msgs := m.Messages
	llog.Debugf("[dst forward队列] 取出当前队列: %v", msgs)
	if len(msgs) == 0 {
		return []Message{}
	}
	m.Messages = []Message{}
	return msgs
}

type DstMsg struct {
	UserName      string `json:"userName"`      // 玩家名称
	SurvivorsName string `json:"survivorsName"` // 角色名称，如 Wendy
	KleiID        string `json:"kleiId"`        // 科雷 id
	Message       string `json:"message"`       // 消息正文
}

func IPWhitelistMiddleware(allowedIPs []string) gin.HandlerFunc {
	llog.Debugf("[dst forward]ip白名单为%s", allowedIPs)
	return func(c *gin.Context) {
		if len(allowedIPs) == 0 {
			llog.Debugf("[dst forward]ip白名单为空，跳过验证")
			c.Next()
			return
		}

		clientIP := c.ClientIP()

		// 检查IP是否在白名单中
		allowed := slices.Contains(allowedIPs, clientIP)

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "访问被拒绝：IP不在白名单中",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (queue *MsgQueue) enqueueGroupMessage(groupMsg *message.GroupMessage) {
	queue.enqueue(Message{
		Type: MsgText,
		Data: Data{
			Source: Source{
				ID:   groupMsg.GroupUin,
				Name: groupMsg.GroupName,
			},
			Sender: Sender{
				ID:   groupMsg.Sender.Uin,
				Name: groupMsg.Sender.CardName,
				Nick: groupMsg.Sender.Nickname,
			},
			Content: groupMsg.ToString(),
		},
	})
}
func (queue *MsgQueue) enqueueCmdMsgByGroup(head string, content any, groupMsg *message.GroupMessage) {
	queue.enqueue(Message{
		Type: MsgCmd,
		Data: Data{
			Source: Source{
				ID:   groupMsg.GroupUin,
				Name: groupMsg.GroupName,
			},
			Sender: Sender{
				ID:   groupMsg.Sender.Uin,
				Name: groupMsg.Sender.CardName,
				Nick: groupMsg.Sender.Nickname,
			},
			Head:    head,
			Content: content,
		},
	})
}
func (queue *MsgQueue) enqueueCmdMsgByPrivate(head string, content any, privateMsg *message.PrivateMessage) {
	queue.enqueue(Message{
		Type: MsgCmd,
		Data: Data{
			Source: Source{
				ID:   privateMsg.Sender.Uin,
				Name: privateMsg.Sender.CardName,
			},
			Sender: Sender{
				ID:   privateMsg.Sender.Uin,
				Name: privateMsg.Sender.CardName,
				Nick: privateMsg.Sender.Nickname,
			},
			Head:    head,
			Content: content,
		},
	})
}

// WriterAdapter 把 io.Writer 的 Write 转发给 llog
type WriterAdapter struct{}
type ErrorWriterAdapter struct{}

func (w *WriterAdapter) Write(p []byte) (n int, err error) {
	llog.Infof("%s", string(p))
	return len(p), nil
}
func (w *ErrorWriterAdapter) Write(p []byte) (n int, err error) {
	llog.Errorf("%s", string(p))
	return len(p), nil
}

func parseDstMsg(m DstMsg) *message.TextElement {
	format := `%s (%s) : %s`
	return message.NewText(fmt.Sprintf(
		format,
		m.UserName,
		m.SurvivorsName,
		m.Message,
	))
}

func initGinWriter() {
	writer := &WriterAdapter{}
	gin.DefaultWriter = writer
	errorWirte := &ErrorWriterAdapter{}
	gin.DefaultErrorWriter = errorWirte
}

var GlobalMsgQueue MsgQueue = DefaultQueue()

func registerServer() {
	initGinWriter()

	otherConfig := config.GlobalConfig.Other

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()

	router.Use(IPWhitelistMiddleware(otherConfig.AllowedIPs))

	router.POST("/send_msg", func(c *gin.Context) {
		var msg DstMsg
		if err := c.ShouldBindJSON(&msg); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		for _, gid := range otherConfig.BindGroups {

			bot.QQClient.Client().SendGroupMessage(
				gid,
				[]message.IMessageElement{parseDstMsg(msg)},
			)
		}

		c.Status(http.StatusOK)
	})

	router.GET("/get_msg", func(c *gin.Context) {
		c.JSON(http.StatusOK, GlobalMsgQueue.drain())
	})

	router.Run(fmt.Sprintf(":%d", otherConfig.GinPort))
}
