package app

import (
	"github.com/LagrangeDev/LagrangeGo/client"
	"github.com/LagrangeDev/LagrangeGo/client/auth"
	"llma.dev/bot"
	"llma.dev/config"
	"llma.dev/logic"
	"llma.dev/utils/llog"
)

// Container 依赖注入容器
type Container struct {
	config       *config.Config
	client       *client.QQClient
	bot          *bot.Bot
	logicManager *logic.LogicManager
}

// NewContainer 创建新的容器实例
func NewContainer() *Container {
	return &Container{}
}

// Initialize 初始化所有依赖
func (c *Container) Initialize() error {
	c.config = &config.Config{}
	// 初始化配置
	config.Init()
	c.config = config.GlobalConfig

	// 初始化日志
	llog.Init(c.config.Log)

	// 创建客户端
	appInfo := auth.AppInfo{
		OS:       "Linux",
		Kernel:   "Linux",
		VendorOS: "linux",

		CurrentVersion:   "3.2.19-39038",
		BuildVersion:     39038,
		MiscBitmap:       32764,
		PTVersion:        "2.0.0",
		PTOSVersion:      19,
		PackageName:      "com.tencent.qq",
		WTLoginSDK:       "nt.wtlogin.0.0.1",
		PackageSign:      "V1_LNX_NQ_3.2.19_39038_GW_B",
		AppID:            1600001615,
		SubAppID:         537258424,
		AppIDQrcode:      13697054,
		AppClientVersion: 39038,

		MainSigmap:  169742560,
		SubSigmap:   0,
		NTLoginType: 1,
	}
	c.client = client.NewClient(c.config.Bot.Account, c.config.Bot.Password)

	// 看LagrangeGo 改不改，改了就用llog，不改就这样适配
	c.client.SetLogger(bot.BotLog{})
	c.client.UseVersion(&appInfo)
	c.client.AddSignServer(c.config.Bot.SignServer)
	c.client.UseDevice(auth.NewDeviceInfo(114514))

	// 创建Bot
	c.bot = bot.NewBot(c.client)
	bot.QQClient = c.bot

	// 加载签名文件
	c.bot.GetAuthManager().LoadSig()

	// 创建逻辑管理器
	c.logicManager = logic.NewLogicManager(c.client)

	// 初始化web 不安全，默认不启用
	// go web.Init(c.bot)
	return nil
}

// GetBot 获取Bot实例
func (c *Container) GetBot() *bot.Bot {
	return c.bot
}

// GetLogicManager 获取逻辑管理器实例
func (c *Container) GetLogicManager() *logic.LogicManager {
	return c.logicManager
}

// GetClient 获取客户端实例
func (c *Container) GetClient() *client.QQClient {
	return c.client
}

// GetConfig 获取配置实例
func (c *Container) GetConfig() *config.Config {
	return c.config
}
