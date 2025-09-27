package dstforward

import (
	"context"
	"strings"

	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/message"
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

	if privateMsg, ok := ctx.GetPrivateMessage(); ok {
		ctx.Client.SendPrivateMessage(privateMsg.Sender.Uin, []message.IMessageElement{
			message.NewText("你说了: " + text),
		})
	} else if groupMsg, ok := ctx.GetGroupMessage(); ok {
		ctx.Client.SendGroupMessage(groupMsg.GroupUin, []message.IMessageElement{
			message.NewText("你说了: " + text),
		})
	}

	return nil
}

// SaveHandler 存档处理器
type SaveHandler struct{}

func (h *SaveHandler) Handle(ctx *logic.MessageContext) error {

	if privateMsg, ok := ctx.GetPrivateMessage(); ok {
		enqueueCmdMsgByPrivate(
			"save",
			privateMsg)
		return nil
	}
	if groupMsg, ok := ctx.GetGroupMessage(); ok {
		enqueueCmdMsgByGroup(
			"save",
			groupMsg)
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

	if privateMsg, ok := ctx.GetPrivateMessage(); ok {
		enqueueCmdMsgByPrivate(
			"rollback",
			privateMsg)
		return nil
	}
	if groupMsg, ok := ctx.GetGroupMessage(); ok {
		enqueueCmdMsgByGroup(
			"rollback",
			groupMsg)
		return nil
	}

	return nil
}

// BanHandler 封禁处理器
type BanHandler struct{}

func (h *BanHandler) Handle(ctx *logic.MessageContext) error {
	if privateMsg, ok := ctx.GetPrivateMessage(); ok {
		enqueueCmdMsgByPrivate(
			"ban",
			privateMsg)
		return nil
	}
	if groupMsg, ok := ctx.GetGroupMessage(); ok {
		enqueueCmdMsgByGroup(
			"ban",
			groupMsg)
		return nil
	}

	return nil
}

// ResetHandler 封禁处理器
type ResetHandler struct{}

func (h *ResetHandler) Handle(ctx *logic.MessageContext) error {
	if privateMsg, ok := ctx.GetPrivateMessage(); ok {
		enqueueCmdMsgByPrivate(
			"reset",
			privateMsg)
		return nil
	}
	if groupMsg, ok := ctx.GetGroupMessage(); ok {
		enqueueCmdMsgByGroup(
			"reset",
			groupMsg)
		return nil
	}

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
/重置世界 - 重置整个世界(谨慎使用)
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
	})
	logic.Manager.HandleCommand("/", "回档", func(ctx *logic.MessageContext) error {
		handler := &RollBackHandler{}
		return handler.Handle(ctx)
	})
	logic.Manager.HandleCommand("/", "重置世界", func(ctx *logic.MessageContext) error {
		handler := &RollBackHandler{}
		return handler.Handle(ctx)
	})
	logic.Manager.HandleCommand("/", "ban", func(ctx *logic.MessageContext) error {
		handler := &BanHandler{}
		return handler.Handle(ctx)
	})

	commands := []string{"/ping", "/help", "/echo"}

	logic.Manager.HandleGroupMessage(func(ctx *logic.MessageContext) error {
		if msg, isOk := ctx.GetGroupMessage(); isOk {
			llog.Debugf("[dst forward]收到群消息:%v", msg)
			msgText := msg.ToString()
			for _, cmd := range commands {
				if strings.HasPrefix(msgText, cmd) {
					return nil
				}
			}
			GlobalMsgQueue.enqueue(Message{Type: 0,
				Data: Data{
					Source: Source{
						ID:   msg.GroupUin,
						Name: msg.GroupName,
					},
					Sender: Sender{
						ID:   msg.Sender.Uin,
						Name: msg.Sender.Nickname,
						Nick: msg.Sender.CardName,
					},
					Content: msgText,
				}})
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
