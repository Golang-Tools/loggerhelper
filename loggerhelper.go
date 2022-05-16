package loggerhelper

import (
	"sync"

	"github.com/Golang-Tools/optparams"
	logrus "github.com/sirupsen/logrus"
)

// 全局变量

var (
	//Logger 默认的logger
	logger *logrus.Logger

	//默认的Log对象
	defaultlog *logrus.Entry

	//锁对象,防止重置出现冲突
	locker *sync.Mutex

	dopts *Options
)

//Dict 简化键值对的写法
type Dict map[string]interface{}

//New 初始化log的配置
func New() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	return log
}

//init 模块初始化
// 赋值默认锁和logger
func init() {
	logger = New()
	locker = &sync.Mutex{}
	_opts := DefaultOpts
	dopts = &_opts
	Set()

}

//GetLogger 获取模块维护得logrus.Logger对象
//该接口用于导出logger给其他模块使用
func GetLogger() *logrus.Logger {
	return logger
}

//Export 导出Log对象
//该函数用于导出当前的默认logrus.Entry给特定模块使用
func Export() *Log {
	if defaultlog == nil {
		return nil
	}
	l := new(Log)
	l.log = defaultlog
	return l
}

//Set 设置logger
//@params opts ...Option 初始化使用的参数,具体可以看options.go文件
func Set(opts ...optparams.Option[Options]) {
	locker.Lock()
	defer locker.Unlock()
	options := optparams.GetOption(dopts, opts...)
	dopts = options
	if options.Output != nil {
		logger.SetOutput(options.Output)
	}
	var fmter logrus.Formatter
	if options.Type == FormatType_Text {
		_fmter := &logrus.TextFormatter{}
		if options.DisableTimeField {
			_fmter.DisableTimestamp = true
		}
		if options.TimeFormat != "" {
			_fmter.TimestampFormat = options.TimeFormat
		}
		if options.DefaultFieldMap != nil {
			_fmter.FieldMap = options.DefaultFieldMap
		}
		fmter = _fmter
	} else {
		_fmter := &logrus.JSONFormatter{}
		if options.DisableTimeField {
			_fmter.DisableTimestamp = true
		}
		if options.TimeFormat != "" {
			_fmter.TimestampFormat = options.TimeFormat
		}
		if options.DefaultFieldMap != nil {
			_fmter.FieldMap = options.DefaultFieldMap
		}
		fmter = _fmter
	}
	logger.SetFormatter(fmter)
	logger.SetLevel(options.Level)
	for _, hook := range options.Hooks {
		logger.Hooks.Add(hook)
	}
	if options.ExtFields == nil || len(options.ExtFields) == 0 {
		defaultlog = nil
		return
	}
	defaultlog = logger.WithFields(options.ExtFields)
}

//Trace 默认log打印Trace级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Trace(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	if defaultlog == nil {
		switch fieldlen {
		case 0:
			{
				logger.Trace(message)
			}
		case 1:
			{
				logger.WithFields(fields[0]).Trace(message)
			}
		default:
			{
				l := logger.WithFields(fields[0])
				for _, field := range fields[1:] {
					l = l.WithFields(field)
				}
				l.Trace(message)
			}
		}
	} else {
		switch fieldlen {
		case 0:
			{
				defaultlog.Trace(message)
			}
		default:
			{
				l := defaultlog
				for _, field := range fields {
					l = l.WithFields(field)
				}
				l.Trace(message)
			}
		}
	}
}

//Debug 默认log打印Debug级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Debug(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	if defaultlog == nil {
		switch fieldlen {
		case 0:
			{
				logger.Debug(message)
			}
		case 1:
			{
				logger.WithFields(fields[0]).Debug(message)
			}
		default:
			{
				l := logger.WithFields(fields[0])
				for _, field := range fields[1:] {
					l = l.WithFields(field)
				}
				l.Debug(message)
			}
		}
	} else {
		switch fieldlen {
		case 0:
			{
				defaultlog.Debug(message)
			}
		default:
			{
				l := defaultlog
				for _, field := range fields {
					l = l.WithFields(field)
				}
				l.Debug(message)
			}
		}
	}
}

//Info 默认log打印Info级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Info(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	if defaultlog == nil {
		switch fieldlen {
		case 0:
			{
				logger.Info(message)
			}
		case 1:
			{
				logger.WithFields(fields[0]).Info(message)
			}
		default:
			{
				l := logger.WithFields(fields[0])
				for _, field := range fields[1:] {
					l = l.WithFields(field)
				}
				l.Info(message)
			}
		}
	} else {
		switch fieldlen {
		case 0:
			{
				defaultlog.Info(message)
			}
		default:
			{
				l := defaultlog
				for _, field := range fields {
					l = l.WithFields(field)
				}
				l.Info(message)
			}
		}
	}
}

//Warn 默认log打印Warn级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Warn(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	if defaultlog == nil {
		switch fieldlen {
		case 0:
			{
				logger.Warn(message)
			}
		case 1:
			{
				logger.WithFields(fields[0]).Warn(message)
			}
		default:
			{
				l := logger.WithFields(fields[0])
				for _, field := range fields[1:] {
					l = l.WithFields(field)
				}
				l.Warn(message)
			}
		}
	} else {
		switch fieldlen {
		case 0:
			{
				defaultlog.Warn(message)
			}
		default:
			{
				l := defaultlog
				for _, field := range fields {
					l = l.WithFields(field)
				}
				l.Warn(message)
			}
		}
	}
}

//Error 默认log打印Error级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Error(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	if defaultlog == nil {
		switch fieldlen {
		case 0:
			{
				logger.Error(message)
			}
		case 1:
			{
				logger.WithFields(fields[0]).Error(message)
			}
		default:
			{
				l := logger.WithFields(fields[0])
				for _, field := range fields[1:] {
					l = l.WithFields(field)
				}
				l.Error(message)
			}
		}
	} else {
		switch fieldlen {
		case 0:
			{
				defaultlog.Error(message)
			}
		default:
			{
				l := defaultlog
				for _, field := range fields {
					l = l.WithFields(field)
				}
				l.Error(message)
			}
		}
	}
}

//Fatal 默认log打印Fatal级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Fatal(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	if defaultlog == nil {
		switch fieldlen {
		case 0:
			{
				logger.Fatal(message)
			}
		case 1:
			{
				logger.WithFields(fields[0]).Fatal(message)
			}
		default:
			{
				l := logger.WithFields(fields[0])
				for _, field := range fields[1:] {
					l = l.WithFields(field)
				}
				l.Fatal(message)
			}
		}
	} else {
		switch fieldlen {
		case 0:
			{
				defaultlog.Fatal(message)
			}
		default:
			{
				l := defaultlog
				for _, field := range fields {
					l = l.WithFields(field)
				}
				l.Fatal(message)
			}
		}
	}
}

//Panic 默认log打印Panic级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Panic(message string, fields ...map[string]interface{}) {
	fieldlen := len(fields)
	if defaultlog == nil {
		switch fieldlen {
		case 0:
			{
				logger.Panic(message)
			}
		case 1:
			{
				logger.WithFields(fields[0]).Panic(message)
			}
		default:
			{
				l := logger.WithFields(fields[0])
				for _, field := range fields[1:] {
					l = l.WithFields(field)
				}
				l.Panic(message)
			}
		}
	} else {
		switch fieldlen {
		case 0:
			{
				defaultlog.Panic(message)
			}
		default:
			{
				l := defaultlog
				for _, field := range fields {
					l = l.WithFields(field)
				}
				l.Panic(message)
			}
		}
	}
}
