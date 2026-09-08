package loggerhelper

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/Golang-Tools/optparams"
)

var (
	//base 当前基础logger(不含扩展字段),通过Set原子替换
	base *slog.Logger

	//curFull 当前基础logger并携带当前扩展字段
	curFull *slog.Logger

	//levelVar 全局共享的等级变量,使已导出的logger等级也能跟随后续Set
	levelVar *slog.LevelVar

	//locker 读写锁,保护 base/curFull/dopts 的并发访问
	locker sync.RWMutex

	dopts *Options

	//loggerhelperPackage 本包路径,用于计算调用方时跳过门面栈帧
	loggerhelperPackage string
)

// Dict 简化键值对的写法
type Dict map[string]interface{}

// New 创建一个独立的默认logger(json格式、debug等级、写入stderr),不与全局状态联动
func New() *slog.Logger {
	return buildBaseLogger(&DefaultOpts, new(slog.LevelVar))
}

// init 模块初始化:解析本包路径、初始化等级变量并应用默认选项
func init() {
	loggerhelperPackage = getPackageName(callerFuncName())
	levelVar = new(slog.LevelVar)
	_opts := DefaultOpts
	dopts = &_opts
	Set()
}

// buildBaseLogger 依据选项构建slog.Logger,handler引用传入的等级变量
func buildBaseLogger(o *Options, lvl *slog.LevelVar) *slog.Logger {
	lvl.Set(o.Level)
	out := o.Output
	if out == nil {
		out = os.Stderr
	}
	hopt := slog.HandlerOptions{
		Level: lvl,
	}
	if o.ReplaceAttr != nil {
		hopt.ReplaceAttr = o.ReplaceAttr
	} else {
		hopt.ReplaceAttr = makeDefaultReplaceAttr(o)
	}
	var h slog.Handler
	if o.Type == FormatType_Text {
		h = slog.NewTextHandler(out, &hopt)
	} else {
		h = slog.NewJSONHandler(out, &hopt)
	}
	for _, mw := range o.HandlerMiddlewares {
		h = mw(h)
	}
	return slog.New(h)
}

// makeDefaultReplaceAttr 内置的v3兼容字段风格:
// message字段名为event、等级输出小写文本、时间字段按选项裁剪或格式化
func makeDefaultReplaceAttr(o *Options) func(groups []string, a slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if len(groups) > 0 {
			return a
		}
		switch a.Key {
		case slog.TimeKey:
			if o.DisableTimeField {
				return slog.Attr{}
			}
			if o.TimeFormat != "" {
				a.Value = slog.StringValue(a.Value.Time().Format(o.TimeFormat))
			}
			return a
		case slog.LevelKey:
			return slog.String(slog.LevelKey, levelName(a.Value.Any().(slog.Level)))
		case slog.MessageKey:
			return slog.String("event", a.Value.String())
		}
		return a
	}
}

// fieldsToArgs 将字段map展开为slog的key-value参数序列
func fieldsToArgs(fields map[string]interface{}) []any {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return args
}

// flatten 依序合并多个字段map为参数序列,后写字段覆盖先写
func flatten(fields []map[string]interface{}) []any {
	if len(fields) == 0 {
		return nil
	}
	if len(fields) == 1 {
		return fieldsToArgs(fields[0])
	}
	m := map[string]interface{}{}
	for _, f := range fields {
		for k, v := range f {
			m[k] = v
		}
	}
	return fieldsToArgs(m)
}

// copyFields 复制字段map,避免与全局配置共享底层数据
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

// GetLogger 获取当前基础logger对象(不含扩展字段)
// 等级阈值跟随后续Set()变化;输出目标与格式等为调用时刻快照
func GetLogger() *slog.Logger {
	locker.RLock()
	defer locker.RUnlock()
	return base
}

// Export 导出Log对象
// 导出的Log固化当前的扩展字段,等级/output等仍跟随后续log.Set设置
func Export() *Log {
	locker.RLock()
	ef := copyFields(dopts.ExtFields)
	locker.RUnlock()
	return &Log{fields: ef}
}

// Set 设置logger
// 每次都构建全新的handler并原子替换base;等级通过共享的levelVar即时生效
// @params opts ...Option 初始化使用的参数,具体可以看options.go文件
func Set(opts ...optparams.Option[Options]) {
	locker.Lock()
	defer locker.Unlock()
	options := optparams.GetOption(dopts, opts...)
	// ExtFields 拷贝为专有副本,避免与调用方传入的map共享导致并发写
	options.ExtFields = copyFields(options.ExtFields)
	b := buildBaseLogger(options, levelVar)
	dopts = options
	base = b
	curFull = b
	if len(options.ExtFields) > 0 {
		curFull = b.With(fieldsToArgs(options.ExtFields)...)
	}
}

// callerFuncName 返回本函数调用者的完整函数名
func callerFuncName() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ""
	}
	return runtime.FuncForPC(pc).Name()
}

// getPackageName 从完整函数名中提取包路径
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

// findCaller 向上查找第一个不属于本包与slog的调用栈帧
func findCaller() *runtime.Frame {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(3, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		pkg := getPackageName(frame.Function)
		if pkg != loggerhelperPackage && pkg != "log/slog" {
			f := frame
			return &f
		}
		if !more {
			break
		}
	}
	return nil
}

// callerAttrs 将业务调用点转为caller/file两个字段
func callerAttrs() []any {
	f := findCaller()
	if f == nil {
		return nil
	}
	return []any{"caller", f.Function, "file", fmt.Sprintf("%s:%d", f.File, f.Line)}
}

// emit 记录一条日志,report为真时附加调用方字段
func emit(l *slog.Logger, report bool, level slog.Level, message string, fields []map[string]interface{}) {
	args := flatten(fields)
	if report {
		args = append(callerAttrs(), args...)
	}
	l.Log(context.Background(), level, message, args...)
}

// output 使用当前全局logger在指定等级输出消息
func output(level slog.Level, message string, fields []map[string]interface{}) {
	locker.RLock()
	full, rc := curFull, dopts.ReportCaller
	locker.RUnlock()
	emit(full, rc, level, message, fields)
}

// Trace 默认log打印Trace级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func Trace(message string, fields ...map[string]interface{}) {
	output(LevelTrace, message, fields)
}

// Debug 默认log打印Debug级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func Debug(message string, fields ...map[string]interface{}) {
	output(slog.LevelDebug, message, fields)
}

// Info 默认log打印Info级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func Info(message string, fields ...map[string]interface{}) {
	output(slog.LevelInfo, message, fields)
}

// Warn 默认log打印Warn级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func Warn(message string, fields ...map[string]interface{}) {
	output(slog.LevelWarn, message, fields)
}

// Error 默认log打印Error级别信息
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func Error(message string, fields ...map[string]interface{}) {
	output(slog.LevelError, message, fields)
}

// Fatal 打印Fatal级别信息并退出进程
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func Fatal(message string, fields ...map[string]interface{}) {
	output(LevelFatal, message, fields)
	os.Exit(1)
}

// Panic 打印Panic级别信息并抛出panic
// @params message string 事件消息
// @params fields ...map[string]interface{} 信息字段
func Panic(message string, fields ...map[string]interface{}) {
	output(LevelPanic, message, fields)
	panic(message)
}
