package loggerhelper

import (
	"io"
	"time"

	"github.com/Golang-Tools/optparams"
	logrus "github.com/sirupsen/logrus"
)

// redis类型
type FormatType int32

const (
	FormatType_JSON FormatType = 0
	FormatType_Text FormatType = 1
)

//Option 设置key行为的选项
//@attribute MaxTTL time.Duration 为0则不设置过期
//@attribute AutoRefresh string 需要为crontab格式的字符串,否则不会自动定时刷新
type Options struct {
	Type             FormatType
	DisableTimeField bool
	TimeFormat       string
	Level            logrus.Level
	ReportCaller     bool
	DefaultFieldMap  logrus.FieldMap
	ExtFields        map[string]interface{}
	Output           io.Writer
	Hooks            []logrus.Hook
}

var DefaultOpts = Options{
	Type:       FormatType_JSON,
	TimeFormat: time.RFC3339Nano,
	Level:      logrus.DebugLevel,
	DefaultFieldMap: logrus.FieldMap{
		logrus.FieldKeyTime:        "time",
		logrus.FieldKeyLevel:       "level",
		logrus.FieldKeyMsg:         "event",
		logrus.FieldKeyLogrusError: "logrus_error",
		logrus.FieldKeyFunc:        "caller",
		logrus.FieldKeyFile:        "file",
	},
	ExtFields: map[string]interface{}{},
	Hooks:     []logrus.Hook{},
}

//WithTextFormat SetLogger函数的参数,用于设置使用text格式替换json格式
func WithTextFormat() optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.Type = FormatType_Text
	})
}

//WithDisableTimeField SetLogger函数的参数,用于设置使用text格式替换json格式
func WithDisableTimeField() optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.DisableTimeField = true
	})
}

//WithTimeFormat SetLogger函数的参数,用于设置使用指定的时间解析格式,默认为RFC3339Nano
func WithTimeFormat(TimeFormat string) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.TimeFormat = TimeFormat
	})
}

//WithLevel SetLogger函数的参数,用于设置log等级
//未知的等级字符串会被忽略并保持原有等级不变
func WithLevel(loglevel string) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		level, err := logrus.ParseLevel(loglevel)
		if err != nil {
			return
		}
		o.Level = level
	})
}

//WithReportCaller SetLogger函数的参数,用于开启调用方信息(caller/file)的输出
func WithReportCaller() optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.ReportCaller = true
	})
}

//AddHooks SetLogger函数的参数,用于增加钩子
func AddHooks(hooks ...logrus.Hook) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = []logrus.Hook{}
		}
		o.Hooks = append(o.Hooks, hooks...)
	})
}

//WithDefaultFieldMap SetLogger函数的参数,用于设置默认字段的新命名
func WithDefaultFieldMap(fm logrus.FieldMap) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.DefaultFieldMap = fm
	})
}

//AddExtField SetLogger函数的参数,用于增加扩展字段
func AddExtField(field string, value interface{}) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		// 复制到新map,避免污染 GetOption 浅拷贝下共享的默认map
		nm := make(map[string]interface{}, len(o.ExtFields)+1)
		for k, v := range o.ExtFields {
			nm[k] = v
		}
		nm[field] = value
		o.ExtFields = nm
	})
}

//WithExtFields SetLogger函数的参数,用于重置扩展字段
func WithExtFields(extFields map[string]interface{}) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.ExtFields = extFields
	})
}

//WithAddExtFields SetLogger函数的参数,用于添加设置扩展字段
func WithAddExtFields(extFields map[string]interface{}) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		// 复制到新map,避免污染 GetOption 浅拷贝下共享的默认map
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

//WithOutput SetLogger函数的参数,用于设置log的写入io
func WithOutput(writer io.Writer) optparams.Option[Options] {
	return optparams.NewFuncOption(func(o *Options) {
		o.Output = writer
	})
}
