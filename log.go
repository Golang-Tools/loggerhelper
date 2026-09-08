package loggerhelper

import logrus "github.com/sirupsen/logrus"

//Log 针对不同业务可以设置不同的Log对象
//它固化了一组字段,其余属性(level/output/hooks等)跟随全局log.Set设置
type Log struct {
	fields map[string]interface{}
}

//entry 基于当前全局logger与该对象固化字段动态构造entry
//每次调用都解析当前logger,使Log跟随后续Set()对其他属性的修改
func (lg *Log) entry() *logrus.Entry {
	locker.RLock()
	l := logger
	locker.RUnlock()
	if len(lg.fields) > 0 {
		return l.WithFields(lg.fields)
	}
	return logrus.NewEntry(l)
}

//GetLogger 获取基于当前logger与该固化字段的logrus.Entry对象
//该接口用于导出logger给其他模块使用,反映调用时刻的全局配置
func (lg *Log) GetLogger() *logrus.Entry {
	return lg.entry()
}

//Trace 打印Trace级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Trace(message string, fields ...map[string]interface{}) {
	emit(lg.entry(), logrus.TraceLevel, message, fields)
}

//Debug 打印Debug级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Debug(message string, fields ...map[string]interface{}) {
	emit(lg.entry(), logrus.DebugLevel, message, fields)
}

//Info 打印Info级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Info(message string, fields ...map[string]interface{}) {
	emit(lg.entry(), logrus.InfoLevel, message, fields)
}

//Warn 打印Warn级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Warn(message string, fields ...map[string]interface{}) {
	emit(lg.entry(), logrus.WarnLevel, message, fields)
}

//Error 打印Error级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Error(message string, fields ...map[string]interface{}) {
	emit(lg.entry(), logrus.ErrorLevel, message, fields)
}

//Fatal 打印Fatal级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Fatal(message string, fields ...map[string]interface{}) {
	emit(lg.entry(), logrus.FatalLevel, message, fields)
}

//Panic 打印Panic级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Panic(message string, fields ...map[string]interface{}) {
	emit(lg.entry(), logrus.PanicLevel, message, fields)
}
