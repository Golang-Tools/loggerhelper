package loggerhelper

import logrus "github.com/sirupsen/logrus"

//针对不同业务可以设置不同的Log对象
//它是固定了一些字段的logger
type Log struct {
	log *logrus.Entry
}

//GetLogger 获取模块维护得logrus.Logger对象
//该接口用于导出logger给其他模块使用
func (lg *Log) GetLogger() *logrus.Entry {
	return lg.log
}

//Trace 打印Trace级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Trace(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	switch fieldlen {
	case 0:
		{
			lg.log.Trace(message)
		}
	default:
		{
			l := lg.log
			for _, field := range fields {
				l = l.WithFields(field)
			}
			l.Trace(message)
		}
	}

}

//Debug 打印Debug级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Debug(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	switch fieldlen {
	case 0:
		{
			lg.log.Debug(message)
		}
	default:
		{
			l := lg.log
			for _, field := range fields {
				l = l.WithFields(field)
			}
			l.Debug(message)
		}
	}

}

//Info 打印Info级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Info(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	switch fieldlen {
	case 0:
		{
			lg.log.Info(message)
		}
	default:
		{
			l := lg.log
			for _, field := range fields {
				l = l.WithFields(field)
			}
			l.Info(message)
		}
	}

}

//Warn 打印Warn级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Warn(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	switch fieldlen {
	case 0:
		{
			lg.log.Warn(message)
		}
	default:
		{
			l := lg.log
			for _, field := range fields {
				l = l.WithFields(field)
			}
			l.Warn(message)
		}
	}

}

//Error 打印Error级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Error(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	switch fieldlen {
	case 0:
		{
			lg.log.Error(message)
		}
	default:
		{
			l := lg.log
			for _, field := range fields {
				l = l.WithFields(field)
			}
			l.Error(message)
		}
	}

}

//Fatal 打印Fatal级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Fatal(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	switch fieldlen {
	case 0:
		{
			lg.log.Fatal(message)
		}
	default:
		{
			l := lg.log
			for _, field := range fields {
				l = l.WithFields(field)
			}
			l.Fatal(message)
		}
	}
}

//Panic 打印Panic级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func (lg *Log) Panic(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	switch fieldlen {
	case 0:
		{
			lg.log.Panic(message)
		}
	default:
		{
			l := lg.log
			for _, field := range fields {
				l = l.WithFields(field)
			}
			l.Panic(message)
		}
	}

}
