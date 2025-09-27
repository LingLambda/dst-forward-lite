package config

import (
	"log"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Bot   BotConfig
	Log   LogConfig
	Other OtherConfig
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
	QrCodePath string   `toml:"qrCodePath"`
	GinPort    uint     `toml:"ginPort"`
	AllowedIPs []string `toml:"allowedIPs"`
}

// GlobalConfig 默认全局配置
var GlobalConfig *Config

// Init 使用 ./application.toml 初始化全局配置
func Init() {
	GlobalConfig = &Config{}
	_, err := toml.DecodeFile("application.toml", GlobalConfig)
	if err != nil {
		log.Panicf("unable to read global config: %v", err)
	}
}

// InitWithContent 从字节数组中读取配置内容
func InitWithContent(configTOMLContent []byte) {
	_, err := toml.Decode(string(configTOMLContent), GlobalConfig)
	if err != nil {
		panic(err)
	}
}
