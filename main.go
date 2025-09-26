package main

import (
	"os"
	"os/signal"
	"syscall"

	"llma.dev/app"
	"llma.dev/logic"
	"llma.dev/utils/llog"
)

func main() {
	// 使用依赖注入容器
	container := app.NewContainer()
	err := container.Initialize()
	if err != nil {
		panic(err)
	}

	bot := container.GetBot()
	logicManager := container.GetLogicManager()

	// 登录
	err = bot.Login()
	if err != nil {
		bot.Client().Release()
		bot.RemoveSig()
		llog.Errorf("[main.初始化] 登录失败，已删除签名，等待用户登录: %s", err)
	} else {
		// 监听
		bot.Listen()

		// 注册自定义逻辑
		logic.Manager = logicManager

		// 设置事件监听
		logicManager.SetupEventListeners()
		defer bot.Client().Release()
		defer bot.Dumpsig()
	}
	// setup the main stop channel
	mc := make(chan os.Signal, 2)
	signal.Notify(mc, os.Interrupt, syscall.SIGTERM)
	for {
		switch <-mc {
		case os.Interrupt, syscall.SIGTERM:
			return
		}
	}
}
