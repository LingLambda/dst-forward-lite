package dstforward

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/message"
	"llma.dev/config"
	"llma.dev/logic"
	"llma.dev/utils/llog"
)

// EchoHandler 回声处理器
type EchoHandler struct{}

func (h *EchoHandler) Handle(ctx *logic.MessageContext) error {
	text := ctx.GetMessageText()
	if text == "" {
		return nil
	}

	ctx.Reply([]message.IMessageElement{
		message.NewText("你说了: " + text),
	})

	return nil
}

// SaveHandler 存档处理器
type SaveHandler struct{}

func (h *SaveHandler) Handle(ctx *logic.MessageContext) error {

	if privateMsg, ok := ctx.GetPrivateMessage(); ok {
		GlobalMsgQueue.enqueueCmdMsgByPrivate(
			"save",
			nil,
			privateMsg)
		ctx.Reply(simpleTextElements("保存成功!"))
		return nil
	}
	if groupMsg, ok := ctx.GetGroupMessage(); ok {
		GlobalMsgQueue.enqueueCmdMsgByGroup(
			"save",
			nil,
			groupMsg)
		ctx.Reply(simpleTextElements("保存成功!"))
		return nil
	}

	return nil
}

// RollBackHandler 回档处理器
type RollBackHandler struct{}

func (h *RollBackHandler) Handle(ctx *logic.MessageContext) error {
	text := ctx.GetMessageText()
	if text == "" {
		return nil
	}
	privateMsg, isPrivate := ctx.GetPrivateMessage()
	groupMsg, isGroup := ctx.GetGroupMessage()

	re := regexp.MustCompile(`/回档\s+(\d+)`) // 捕获数字
	match := re.FindStringSubmatch(text)
	var dayNum int
	if len(match) > 1 {
		dayNum, _ = strconv.Atoi(match[1])
		llog.Debugf("[回档]匹配到数字: %d", dayNum)
	}
	if len(match) <= 1 || dayNum < 1 {
		ctx.Reply(simpleTextElements("请输入有效的回档天数 示例: /回档 1"))
		return nil
	} else if dayNum > 100 {
		ctx.Reply(simpleTextElements("输入的回档天数过长，请输入有效的回档天数 示例: /回档 1"))
		return nil
	}

	confirmFunc := func() {
		okElements := simpleTextElements(fmt.Sprintf("已下发回档 %d 天命令", dayNum))
		if isPrivate {
			GlobalMsgQueue.enqueueCmdMsgByPrivate(
				"rollback",
				dayNum,
				privateMsg)
			ctx.Reply(okElements)
			return
		}
		if isGroup {
			GlobalMsgQueue.enqueueCmdMsgByGroup(
				"rollback",
				dayNum,
				groupMsg)
			ctx.Reply(okElements)
			return
		}
	}

	cancelFunc := func() {
		ctx.Reply(simpleTextElements("已取消回档操作"))
	}

	ctx.Prompt(fmt.Sprintf("回档 %d 天", dayNum), 30*time.Second, confirmFunc, cancelFunc)
	return nil
}

// BanHandler 封禁处理器
type BanHandler struct{}

func (h *BanHandler) Handle(ctx *logic.MessageContext) error {
	text := ctx.GetMessageText()
	if text == "" {
		return nil
	}

	re := regexp.MustCompile(`/ban\s+(KU_\S+)`) // 捕获字符串
	match := re.FindStringSubmatch(text)
	var kleiId string
	if len(match) > 1 {
		kleiId = match[1]
		llog.Debugf("[ban]匹配到kleiid: %s", kleiId)
	}
	if len(match) <= 1 {
		errorElemets := simpleTextElements("请输入有效的kleiid 示例: /ban KU_xxxxx")
		ctx.Reply(errorElemets)
		return nil
	}

	okElements := simpleTextElements(fmt.Sprintf("已将用户 %s 封禁", kleiId))
	if privateMsg, ok := ctx.GetPrivateMessage(); ok {
		GlobalMsgQueue.enqueueCmdMsgByPrivate(
			"ban",
			kleiId,
			privateMsg)
		ctx.Reply(okElements)
		return nil
	}
	if groupMsg, ok := ctx.GetGroupMessage(); ok {
		GlobalMsgQueue.enqueueCmdMsgByGroup(
			"ban",
			kleiId,
			groupMsg)
		ctx.Reply(okElements)
		return nil
	}

	return nil
}

// ResetHandler 重置世界处理器
type ResetHandler struct{}

func (h *ResetHandler) Handle(ctx *logic.MessageContext) error {

	confirmFunc := func() {
		okElements := simpleTextElements("已重置世界")
		if privateMsg, ok := ctx.GetPrivateMessage(); ok {
			GlobalMsgQueue.enqueueCmdMsgByPrivate(
				"reset",
				nil,
				privateMsg)
			ctx.Reply(okElements)
		}
		if groupMsg, ok := ctx.GetGroupMessage(); ok {
			GlobalMsgQueue.enqueueCmdMsgByGroup(
				"reset",
				nil,
				groupMsg)
			ctx.Reply(okElements)
		}
	}

	cancelFunc := func() {
		ctx.Reply(simpleTextElements("已取消重置世界操作"))
	}

	ctx.Prompt("重置世界", 30*time.Second, confirmFunc, cancelFunc)
	return nil
}

// HelpHandler 帮助命令处理器
type HelpHandler struct{}

func (h *HelpHandler) Handle(ctx *logic.MessageContext) error {
	help := `可用命令:
/ping - 测试连接
/help - 显示帮助
/echo <消息> - 回声消息
/回档 <天数> - 回档指定天数
/保存 - 即时存档
/重置世界 - 重新生成整个世界(谨慎使用)
/ban <科雷id> - 封禁玩家
`

	if privateMsg, ok := ctx.GetPrivateMessage(); ok {
		ctx.Client.SendPrivateMessage(privateMsg.Sender.Uin, []message.IMessageElement{
			message.NewText(help),
		})
	} else if groupMsg, ok := ctx.GetGroupMessage(); ok {
		ctx.Client.SendGroupMessage(groupMsg.GroupUin, []message.IMessageElement{
			message.NewText(help),
		})
	}

	return nil
}

// RegisterCustomLogic 注册所有自定义逻辑
func RegisterCustomLogic() {
	if logic.Manager == nil {
		llog.Errorf("[plugin.gemini_chat] Logiclogic.Manager 未初始化")
		return
	}
	// 身份认证列表
	authMiddle := logic.AuthMiddleware(config.GlobalConfig.Other.AllowedUIDs)

	// 注册help命令
	logic.Manager.HandleCommand("/", "help", func(ctx *logic.MessageContext) error {
		handler := &HelpHandler{}
		return handler.Handle(ctx)
	})

	// 注册echo命令
	logic.Manager.HandleCommand("/", "echo", func(ctx *logic.MessageContext) error {
		handler := &EchoHandler{}
		return handler.Handle(ctx)
	})

	logic.Manager.HandleCommand("/", "保存", func(ctx *logic.MessageContext) error {
		handler := &SaveHandler{}
		return handler.Handle(ctx)
	}, authMiddle)
	logic.Manager.HandleCommand("/", "回档", func(ctx *logic.MessageContext) error {
		handler := &RollBackHandler{}
		return handler.Handle(ctx)
	}, authMiddle)
	logic.Manager.HandleCommand("/", "重置世界", func(ctx *logic.MessageContext) error {
		handler := &ResetHandler{}
		return handler.Handle(ctx)
	}, authMiddle)
	logic.Manager.HandleCommand("/", "ban", func(ctx *logic.MessageContext) error {
		handler := &BanHandler{}
		return handler.Handle(ctx)
	}, authMiddle)

	commands := []string{"/ping", "/help", "/echo", "/回档", "/重置世界", "/ban"}

	// 转发
	logic.Manager.HandleGroupMessage(func(ctx *logic.MessageContext) error {
		if msg, isOk := ctx.GetGroupMessage(); isOk {
			llog.Debugf("[dst forward]收到群消息:%v", msg)
			msgText := msg.ToString()
			for _, cmd := range commands {
				if strings.HasPrefix(msgText, cmd) {
					return nil
				}
			}
			GlobalMsgQueue.enqueueGroupMessage(msg)
		}
		return nil
	})

	// 注册事件监听器
	logic.Manager.GetEventBus().Subscribe(logic.EventTypeCommandExecuted, func(ctx context.Context, event logic.Event) error {
		if msgEvent, ok := event.(*logic.MessageEvent); ok {
			command := msgEvent.MessageContext.GetString("executed_command")
			llog.Infof("[plugin.gemini_chat] 命令 %s 已执行", command)
		}
		return nil
	})

	llog.Infof("[plugin.gemini_chat] 自定义逻辑注册完成")
}

// 向后兼容的处理器实现
type PrivateMessageHandler struct{}

func (h *PrivateMessageHandler) Handle(client *client.QQClient, msg any) error {
	if event, ok := msg.(*message.PrivateMessage); ok {
		client.SendPrivateMessage(event.Sender.Uin, []message.IMessageElement{
			message.NewText("Hello World!"),
		})
	}
	return nil
}

type GroupMessageHandler struct{}

func (h *GroupMessageHandler) Handle(client *client.QQClient, msg any) error {
	if event, ok := msg.(*message.GroupMessage); ok {
		client.SendGroupMessage(event.GroupUin, []message.IMessageElement{
			message.NewText("Hello World!"),
		})
	}
	return nil
}

func simpleTextElements(text string) []message.IMessageElement {
	return []message.IMessageElement{&message.TextElement{Content: text}}
}
