package web

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"llma.dev/bot"
	"llma.dev/config"
	"llma.dev/utils/llog"
)

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

func initGinWriter() {
	writer := &WriterAdapter{}
	gin.DefaultWriter = writer
	errorWirte := &ErrorWriterAdapter{}
	gin.DefaultErrorWriter = errorWirte
}

func Init(client *bot.Bot) {
	initGinWriter()
	router := gin.Default()
	gin.Logger()

	// This handler will match /user/john but will not match /user/ or /user
	router.GET("/bot/status", func(c *gin.Context) {
		var state string
		if client == nil {
			c.String(http.StatusOK, "bot死了")
		}
		s := client.GetState()
		switch s {
		case bot.Disconnected:
			state = "已断开"
		case bot.Connecting:
			state = "连接中"
		case bot.Connected:
			state = "已连接"
		case bot.Reconnecting:
			state = "等待连接"
		default:
			state = "未知状态"
		}
		c.String(http.StatusOK, state)
	})

	router.GET("bot/qrcode", func(c *gin.Context) {
		// 打开本地 PNG 文件
		file, err := os.Open(config.GlobalConfig.Other.QrCodePath)
		if err != nil {
			c.String(http.StatusInternalServerError, "无法打开文件")
			return
		}
		defer file.Close()

		// 使用 io.Copy 把文件内容写入响应
		if _, err := io.Copy(c.Writer, file); err != nil {
			c.String(http.StatusInternalServerError, "写入响应失败")
			return
		}
	})

	router.GET("bot/kill", func(c *gin.Context) {
		client := bot.QQClient
		if client == nil {
			c.String(http.StatusOK, "bot早就死了")
			return
		}
		client.Stop()
		client.RemoveSig()
		c.String(http.StatusOK, "一破，卧龙出山")
	})

	router.GET("bot/login", func(c *gin.Context) {
		client := bot.QQClient
		if client != nil {
			c.String(http.StatusOK, "bot还活着，先杀了再来")
			return
		}
		client.Login()
		c.String(http.StatusOK, "OK")
	})

	router.Run(fmt.Sprintf(":%d", config.GlobalConfig.Other.GinPort))
}
