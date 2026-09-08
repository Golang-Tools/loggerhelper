package loggerhelper

import (
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/Golang-Tools/optparams"
)

// FormatType 输出格式类型
type FormatType int32

const (
	//FormatType_JSON json格式
	FormatType_JSON FormatType = 0
	//FormatType_Text text格式
	FormatType_Text FormatType = 1
)

// 自定义的日志级别,slog 未提供 trace/fatal/panic
const (
	//LevelTrace 低于 debug 的追踪级
	LevelTrace slog.Level = slog.LevelDebug - 4
	//LevelFatal 高于 error 的致命级
	LevelFatal slog.Level = slog.LevelError + 4
	//LevelPanic 最高的崩溃级
	LevelPanic slog.Level = slog.LevelError + 8
)

// HandlerMiddleware 包装一个 slog.Handler 返回新的 handler,用于扩展日志处理
type HandlerMiddleware func(slog.Handler) slog.Handler

// Options 设置logger行为的选项
type Options struct {
	Type             FormatType
	DisableTimeField bool
	TimeFormat       string
	Level            slog.Level
	ReportCaller     bool
	//ReplaceAttr 完全接管字段的改写(如等级小写化、消息字段改名);为空则使用内置的v3兼容风格
	ReplaceAttr        func(groups []string, a slog.Attr) slog.Attr
	ExtFields          map[string]interface{}
	Output             io.Writer
	HandlerMiddlewares []HandlerMiddleware
}

// DefaultOpts 默认选项
var DefaultOpts = Options{
	Type:       FormatType_JSON,
	TimeFormat: time.RFC3339Nano,
	Level:      slog.LevelDebug,
	ExtFields:  map[string]interface{}{},
}

// WithTextFormat 设置使用text格式替换json格式
func WithTextFormat() optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.Type = FormatType_Text
	})
}

// WithJSONFormat 设置使用json格式(默认),可将text格式切换回json
func WithJSONFormat() optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.Type = FormatType_JSON
	})
}

// WithDisableTimeField 关闭时间字段的输出
func WithDisableTimeField() optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.DisableTimeField = true
	})
}

// WithTimeFormat 设置时间字段的解析格式,默认RFC3339Nano
func WithTimeFormat(TimeFormat string) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.TimeFormat = TimeFormat
	})
}

// WithLevel 设置日志等级,支持 trace/debug/info/warn(warning)/error/fatal/panic(大小写不敏感)
// 未知的等级字符串会被忽略并保持原有等级不变
func WithLevel(level string) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		if l, ok := parseLevel(level); ok {
			o.Level = l
		}
	})
}

// WithReportCaller 开启调用方信息输出(caller/file字段指向业务调用处)
func WithReportCaller() optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.ReportCaller = true
	})
}

// WithDisableReportCaller 关闭调用方信息输出
func WithDisableReportCaller() optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.ReportCaller = false
	})
}

// WithReplaceAttr 用自定义函数完全接管字段改写
// 若不设置,则使用内置的v3兼容风格:message字段名为event、等级输出小写文本
func WithReplaceAttr(fn func(groups []string, a slog.Attr) slog.Attr) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.ReplaceAttr = fn
	})
}

// WithOutput 设置日志的写入io,为空时默认写入os.Stderr
func WithOutput(writer io.Writer) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.Output = writer
	})
}

// WithExtFields 重置扩展字段
func WithExtFields(extFields map[string]interface{}) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.ExtFields = extFields
	})
}

// AddExtField 增加单个扩展字段
func AddExtField(field string, value interface{}) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		// 复制到新map,避免污染浅拷贝下共享的默认map
		nm := make(map[string]interface{}, len(o.ExtFields)+1)
		for k, v := range o.ExtFields {
			nm[k] = v
		}
		nm[field] = value
		o.ExtFields = nm
	})
}

// WithAddExtFields 追加扩展字段
func WithAddExtFields(extFields map[string]interface{}) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		// 复制到新map,避免污染浅拷贝下共享的默认map
		nm := make(map[string]interface{}, len(o.ExtFields)+len(extFields))
		for k, v := range o.ExtFields {
			nm[k] = v
		}
		for k, v := range extFields {
			nm[k] = v
		}
		o.ExtFields = nm
	})
}

// parseLevel 将字符串转为slog.Level
func parseLevel(level string) (slog.Level, bool) {
	switch strings.ToLower(level) {
	case "trace":
		return LevelTrace, true
	case "debug":
		return slog.LevelDebug, true
	case "info":
		return slog.LevelInfo, true
	case "warn", "warning":
		return slog.LevelWarn, true
	case "error":
		return slog.LevelError, true
	case "fatal":
		return LevelFatal, true
	case "panic":
		return LevelPanic, true
	}
	return 0, false
}

// levelName 将级别转为输出用的小写文本(与v3风格一致,warn输出为warning)
func levelName(l slog.Level) string {
	switch {
	case l <= LevelTrace:
		return "trace"
	case l <= slog.LevelDebug:
		return "debug"
	case l <= slog.LevelInfo:
		return "info"
	case l <= slog.LevelWarn:
		return "warning"
	case l <= slog.LevelError:
		return "error"
	case l <= LevelFatal:
		return "fatal"
	default:
		return "panic"
	}
}
