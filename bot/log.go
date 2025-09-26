package bot

import "llma.dev/utils/llog"

// lagrange Log 适配接口
type BotLog struct{}

func (l BotLog) Info(format string, args ...any) {
	llog.Infof(format, args...)
}
func (l BotLog) Warning(format string, args ...any) {
	llog.Warningf(format, args...)
}
func (l BotLog) Error(format string, args ...any) {
	llog.Errorf(format, args...)
}
func (l BotLog) Debug(format string, args ...any) {
	llog.Debugf(format, args...)
}
func (l BotLog) Dump(dumped []byte, format string, args ...any) {
	llog.Dump(dumped, format, args...)
}
