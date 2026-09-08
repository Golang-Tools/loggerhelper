package loggerhelper

import (
	"runtime"
	"strings"
	"sync"

	"github.com/Golang-Tools/optparams"
	logrus "github.com/sirupsen/logrus"
)

// 全局变量

var (
	//logger 当前的logger,通过Set原子替换
	logger *logrus.Logger

	//locker 读写锁,保护 logger/dopts 的并发访问
	locker sync.RWMutex

	dopts *Options
)

//Dict 简化键值对的写法
type Dict map[string]interface{}

//logrusPackage logrus 的包路径,用于计算调用方时跳过其栈帧
const logrusPackage = "github.com/sirupsen/logrus"

//loggerhelperPackage 本包路径,用于计算调用方时跳过门面栈帧
var loggerhelperPackage string

//New 初始化log的配置
func New() *logrus.Logger {
	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})
	return log
}

//init 模块初始化
// 赋值默认logger并解析出本包路径
func init() {
	loggerhelperPackage = getPackageName(callerFuncName())
	logger = New()
	_opts := DefaultOpts
	dopts = &_opts
	Set()
}

//callerFuncName 返回本函数调用者的完整函数名
func callerFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	return runtime.FuncForPC(pc).Name()
}

//getPackageName 从完整函数名中提取包路径
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}
	return f
}

//callerHook 修正调用方信息,使其指向业务代码而非本门面包
type callerHook struct{}

func (h callerHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

func (h callerHook) Fire(entry *logrus.Entry) error {
	if c := findCaller(); c != nil {
		entry.Caller = c
	}
	return nil
}

//findCaller 向上查找第一个既不属于 logrus 也不属于本包的调用栈帧
func findCaller() *runtime.Frame {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		pkg := getPackageName(frame.Function)
		if pkg != loggerhelperPackage && pkg != logrusPackage {
			f := frame
			return &f
		}
		if !more {
			break
		}
	}
	return nil
}

//GetLogger 获取模块维护得logrus.Logger对象
//该接口用于导出logger给其他模块使用
func GetLogger() *logrus.Logger {
	locker.RLock()
	defer locker.RUnlock()
	return logger
}

//Export 导出Log对象
//导出的Log固化当前的扩展字段,但等级/output等仍跟随后续log.Set设置
func Export() *Log {
	locker.RLock()
	ef := copyFields(dopts.ExtFields)
	locker.RUnlock()
	return &Log{fields: ef}
}

//copyFields 复制字段map,避免与全局配置共享底层数据
func copyFields(src map[string]interface{}) map[string]interface{} {
	if len(src) == 0 {
		return nil
	}
	cp := make(map[string]interface{}, len(src))
	for k, v := range src {
		cp[k] = v
	}
	return cp
}

//buildFormatter 根据选项构造 logrus.Formatter
func buildFormatter(o *Options) logrus.Formatter {
	if o.Type == FormatType_Text {
		f := &logrus.TextFormatter{}
		f.DisableTimestamp = o.DisableTimeField
		if o.TimeFormat != "" {
			f.TimestampFormat = o.TimeFormat
		}
		if o.DefaultFieldMap != nil {
			f.FieldMap = o.DefaultFieldMap
		}
		return f
	}
	f := &logrus.JSONFormatter{}
	f.DisableTimestamp = o.DisableTimeField
	if o.TimeFormat != "" {
		f.TimestampFormat = o.TimeFormat
	}
	if o.DefaultFieldMap != nil {
		f.FieldMap = o.DefaultFieldMap
	}
	return f
}

//Set 设置logger
//每次都构建一个全新的logger再原子替换,已发布的logger不再被修改,
//从而在logrus不对Formatter加锁的前提下,保证与并发日志的读写安全
//@params opts ...Option 初始化使用的参数,具体可以看options.go文件
func Set(opts ...optparams.Option[Options]) {
	locker.Lock()
	defer locker.Unlock()
	options := optparams.GetOption(dopts, opts...)
	newLogger := logrus.New()
	if options.Output != nil {
		newLogger.SetOutput(options.Output)
	}
	newLogger.SetFormatter(buildFormatter(options))
	newLogger.SetLevel(options.Level)
	hooks := logrus.LevelHooks{}
	for _, hook := range options.Hooks {
		hooks.Add(hook)
	}
	newLogger.SetReportCaller(options.ReportCaller)
	if options.ReportCaller {
		hooks.Add(callerHook{})
	}
	newLogger.ReplaceHooks(hooks)
	// ExtFields 拷贝为专有副本,避免与调用方传入的map共享导致并发写
	options.ExtFields = copyFields(options.ExtFields)
	logger = newLogger
	dopts = options
}

//emit 将字段应用到 entry 后按等级输出,保留 Fatal 退出与 Panic 抛出的语义
func emit(e *logrus.Entry, level logrus.Level, message string, fields []map[string]interface{}) {
	for _, field := range fields {
		e = e.WithFields(field)
	}
	switch level {
	case logrus.TraceLevel:
		e.Trace(message)
	case logrus.DebugLevel:
		e.Debug(message)
	case logrus.InfoLevel:
		e.Info(message)
	case logrus.WarnLevel:
		e.Warn(message)
	case logrus.ErrorLevel:
		e.Error(message)
	case logrus.FatalLevel:
		e.Fatal(message)
	case logrus.PanicLevel:
		e.Panic(message)
	}
}

//output 使用默认log在指定等级输出消息
//每次调用解析当前logger与扩展字段,保持与Set的并发安全
func output(level logrus.Level, message string, fields []map[string]interface{}) {
	locker.RLock()
	lg, ef := logger, dopts.ExtFields
	locker.RUnlock()
	e := logrus.NewEntry(lg)
	if len(ef) > 0 {
		e = e.WithFields(ef)
	}
	emit(e, level, message, fields)
}

//Trace 默认log打印Trace级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Trace(message string, fields ...map[string]interface{}) {
	output(logrus.TraceLevel, message, fields)
}

//Debug 默认log打印Debug级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Debug(message string, fields ...map[string]interface{}) {
	output(logrus.DebugLevel, message, fields)
}

//Info 默认log打印Info级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Info(message string, fields ...map[string]interface{}) {
	output(logrus.InfoLevel, message, fields)
}

//Warn 默认log打印Warn级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Warn(message string, fields ...map[string]interface{}) {
	output(logrus.WarnLevel, message, fields)
}

//Error 默认log打印Error级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Error(message string, fields ...map[string]interface{}) {
	output(logrus.ErrorLevel, message, fields)
}

//Fatal 默认log打印Fatal级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Fatal(message string, fields ...map[string]interface{}) {
	output(logrus.FatalLevel, message, fields)
}

//Panic 默认log打印Panic级别信息
//@params message string 事件消息
//@params fields ...map[string]interface{} 信息字段
func Panic(message string, fields ...map[string]interface{}) {
	output(logrus.PanicLevel, message, fields)
}
