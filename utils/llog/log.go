package llog

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"llma.dev/config"
)

type Logger struct {
	log *logrus.Logger
}

var Log *Logger

func DefaultLogConfig() *config.LogConfig {
	return &config.LogConfig{
		Level:      "info",    // 默认 info 级别
		EnableFile: false,     // 默认不输出到文件
		FilePath:   "app.log", // 默认日志文件路径
		MaxSize:    100,       // 默认单个日志文件最大 100 MB
		MaxBackups: 7,         // 默认保留 7 个旧日志
		MaxAge:     30,        // 默认保留 30 天旧日志
		Format:     "text",    // 默认文本格式输出
	}
}
func Init(cfg config.LogConfig) {
	Log = &Logger{
		log: logrus.New(),
	}
	logger := Log.log

	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// 设置日志格式
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&MyFormatter{})
	}

	logOutput := &lumberjack.Logger{
		Filename:   cfg.FilePath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   true,
	}

	// 设置输出
	if cfg.EnableFile {
		logger.SetOutput(io.MultiWriter(logOutput, os.Stdout))
	} else {
		logger.SetOutput(os.Stdout)
	}
}

type MyFormatter struct{}

// 将Level限制为4字符长度
func fixed4Upper(s string) string {
	rs := []rune(strings.ToUpper(s))
	if len(rs) > 4 {
		rs = rs[:4]
	}
	return fmt.Sprintf("%-4s", string(rs))
}

// 实现Format方法
func (f *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	logLine := fmt.Sprintf("[%s] [%s]: %s",
		entry.Time.Format("2006-01-02 15:04:05"), // 时间
		fixed4Upper(entry.Level.String()),        // 级别
		entry.Message,                            // 消息
	)

	// 附加字段
	for k, v := range entry.Data {
		logLine += fmt.Sprintf(" %s=%v", k, v)
	}
	logLine += "\n"
	return []byte(logLine), nil
}

func (l Logger) Infof(format string, args ...any) {
	l.log.Infof(format, args...)
}

func (l Logger) Warningf(format string, args ...any) {
	l.log.Warningf(format, args...)
}

func (l Logger) Errorf(format string, args ...any) {
	l.log.Errorf(format, args...)
}

func (l Logger) Debugf(format string, args ...any) {
	l.log.Debugf(format, args...)
}

func (l Logger) Fatalf(format string, args ...any) {
	l.log.Fatalf(format, args...)
}

func (l Logger) Info(args ...any) {
	l.log.Info(args...)
}

func (l Logger) Warning(args ...any) {
	l.log.Warning(args...)
}

func (l Logger) Error(args ...any) {
	l.log.Error(args...)
}

// 为了实现lagrange的接口，这里必须是stirng 和 ...any两个形参
func (l Logger) Debug(args ...any) {
	l.log.Debug(args...)
}

func (l Logger) Fatal(args ...any) {
	l.log.Fatal(args...)
}

func (l Logger) Dump(dumped []byte, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	l.log.Infof("%s: %s", msg, string(dumped)) // 直接当字符串打印
}

func (l Logger) WithField(key string, value ...any) Logger {
	return Logger{
		log: l.log.WithField(key, value).Logger,
	}
}

func Infof(format string, args ...any) {
	Log.Infof(format, args...)
}

func Warningf(format string, args ...any) {
	Log.Warningf(format, args...)

}
func Errorf(format string, args ...any) {
	Log.Errorf(format, args...)

}
func Debugf(format string, args ...any) {
	Log.Debugf(format, args...)

}
func Fatalf(format string, args ...any) {
	Log.Fatalf(format, args...)

}

func Dump(dumped []byte, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	Log.Infof("%s: %s", msg, string(dumped))
}

func GetLogrus() *logrus.Logger {
	return Log.log
}
