package config

import (
	"log"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Bot   BotConfig   `toml:"bot"`
	Log   LogConfig   `toml:"log"`
	Other OtherConfig `toml:"other"`
}

// BotConfig 代表TOML文件中的bot部分
type BotConfig struct {
	Account    uint32 `toml:"account"`
	Password   string `toml:"password"`
	SignServer string `toml:"signServer"`
}
type LogConfig struct {
	Level      string `toml:"level"`      // 日志级别: debug, info, warn, error
	EnableFile bool   `toml:"enableFile"` // 是否启用文件输出
	FilePath   string `toml:"filePath"`   // 日志文件路径
	MaxSize    int    `toml:"maxSize"`    // 单个日志文件最大大小(MB)
	MaxBackups int    `toml:"maxBackups"` // 保留的旧文件个数
	MaxAge     int    `toml:"maxAge"`     // 保留的旧文件天数
	Format     string `toml:"format"`     // 输出格式: text, json
}
type OtherConfig struct {
	QrCodePath    string   `toml:"qrCodePath"`
	GinPort       uint     `toml:"ginPort"`
	AllowedIPs    []string `toml:"allowedIPs"`
	AllowedUIDs   []uint32 `toml:"allowedUIDs"`
	AllowedGroups []uint32 `toml:"allowedGroups"`
	BindGroups    []uint32 `toml:"bindGroups"`
}

// 配置文件名
const FILE_NAME string = "application.toml"

// GlobalConfig 默认全局配置
var GlobalConfig *Config

// Init 使用本地toml文件初始化全局配置
func Init() {
	GlobalConfig = &Config{}

	checkDefaultConfigFile()

	_, err := toml.DecodeFile(FILE_NAME, GlobalConfig)
	if err != nil {
		log.Printf("读取配置文件 %s 错误，请检查配置文件语法是否正确: %v", FILE_NAME, err)
	}
	if len(GlobalConfig.Other.BindGroups) == 0 {
		log.Printf("%s!!!警告!!!:您未在 %s 配置任何绑定群聊，饥荒联机版的消息将不会被转发！！%s", "\033[31m", FILE_NAME, "\033[0m")
	}
}

// 检查配置文件若没有则创建
func checkDefaultConfigFile() {
	_, err := os.Stat(FILE_NAME)
	if err != nil {
		log.Printf("配置文件 %s 不存在，正在创建默认配置文件", FILE_NAME)

		file, err := os.Create(FILE_NAME)
		if err != nil {
			panic(err)
		}
		defer file.Close()

		encoder := toml.NewEncoder(file)
		err = encoder.Encode(DefaultConfig())
		if err != nil {
			log.Panicf("创建配置文件 %s 失败", FILE_NAME)
		}
		log.Printf("创建配置文件 %s 成功，请在程序目录下的配置文件中填写账号密码后重新启动程序", FILE_NAME)
		os.Exit(0)
	}
}

func DefaultConfig() Config {
	bot := BotConfig{
		Account:    0,
		Password:   "111111",
		SignServer: "https://sign.lagrangecore.org/api/sign/39038",
	}
	log := LogConfig{
		Level:      "info",
		EnableFile: true,
		FilePath:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 20,
		MaxAge:     30,
		Format:     "text",
	}
	other := OtherConfig{
		QrCodePath: "crcode.png",
		GinPort:    5562,
		AllowedIPs: []string{
			"127.0.0.1",
			"192.168.1.100",
			"10.0.0.1",
		},
		AllowedUIDs:   []uint32{},
		AllowedGroups: []uint32{},
		BindGroups:    []uint32{},
	}

	return Config{
		Bot:   bot,
		Log:   log,
		Other: other,
	}
}

// InitWithContent 从字节数组中读取配置内容
func InitWithContent(configTOMLContent []byte) {
	_, err := toml.Decode(string(configTOMLContent), GlobalConfig)
	if err != nil {
		panic(err)
	}
}
