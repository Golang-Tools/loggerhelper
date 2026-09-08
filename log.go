package loggerhelper

import (
	"log/slog"
	"os"
)

// Log 针对不同业务可以设置不同的Log对象
// 它固化了一组字段,其余属性(等级/output等)跟随全局log.Set设置
type Log struct {
	fields map[string]interface{}
}

// GetLogger 获取基于当前全局logger并携带该对象固化字段的slog.Logger
func (lg *Log) GetLogger() *slog.Logger {
	return lg.current().With(fieldsToArgs(lg.fields)...)
}

// current 当前全局基础logger与调用方开关的快照
func (lg *Log) current() *slog.Logger {
	locker.RLock()
	b := base
	locker.RUnlock()
	return b
}

// log 使用当前全局logger在指定等级输出,附带固化字段
func (lg *Log) log(level slog.Level, message string, fields []map[string]interface{}) {
	locker.RLock()
	rc := dopts.ReportCaller
	locker.RUnlock()
	l := lg.current()
	if len(lg.fields) > 0 {
		l = l.With(fieldsToArgs(lg.fields)...)
	}
	emit(l, rc, level, message, fields)
}

// Trace 打印Trace级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func (lg *Log) Trace(message string, fields ...map[string]interface{}) {
	lg.log(LevelTrace, message, fields)
}

// Debug 打印Debug级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func (lg *Log) Debug(message string, fields ...map[string]interface{}) {
	lg.log(slog.LevelDebug, message, fields)
}

// Info 打印Info级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func (lg *Log) Info(message string, fields ...map[string]interface{}) {
	lg.log(slog.LevelInfo, message, fields)
}

// Warn 打印Warn级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func (lg *Log) Warn(message string, fields ...map[string]interface{}) {
	lg.log(slog.LevelWarn, message, fields)
}

// Error 打印Error级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func (lg *Log) Error(message string, fields ...map[string]interface{}) {
	lg.log(slog.LevelError, message, fields)
}

// Fatal 打印Fatal级别信息并退出进程
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func (lg *Log) Fatal(message string, fields ...map[string]interface{}) {
	lg.log(LevelFatal, message, fields)
	os.Exit(1)
}

// Panic 打印Panic级别信息并抛出panic
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func (lg *Log) Panic(message string, fields ...map[string]interface{}) {
	lg.log(LevelPanic, message, fields)
	panic(message)
}
