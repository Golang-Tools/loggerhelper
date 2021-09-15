package loggerhelper

import (
	"fmt"
	"io"
	"time"

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

// Option configures how we set up the connection.
type Option interface {
	Apply(*Options)
}

// func (emptyOption) apply(*Options) {}
type funcOption struct {
	f func(*Options)
}

func (fo *funcOption) Apply(do *Options) {
	fo.f(do)
}

func newFuncOption(f func(*Options)) *funcOption {
	return &funcOption{
		f: f,
	}
}

//WithTextFormat Init函数的参数,用于设置使用text格式替换json格式
func WithTextFormat() Option {
	return newFuncOption(func(o *Options) {
		o.Type = FormatType_Text
	})
}

//WithDisableTimeField Init函数的参数,用于设置使用text格式替换json格式
func WithDisableTimeField() Option {
	return newFuncOption(func(o *Options) {
		o.DisableTimeField = true
	})
}

//WithTimeFormat Init函数的参数,用于设置使用指定的时间解析格式,默认为RFC3339Nano
func WithTimeFormat(TimeFormat string) Option {
	return newFuncOption(func(o *Options) {
		o.TimeFormat = TimeFormat
	})
}

//parseLevel 将字符串转为`logrus.Level`,未知的字符串默认匹配为"logrus.DebugLevel"
func parseLevel(loglevel string) logrus.Level {
	level, err := logrus.ParseLevel(loglevel)
	if err != nil {
		fmt.Printf("未知的等级`%s`,使用默认值`Debug`\n", loglevel)
		return logrus.DebugLevel
	}
	return level
}

//WithLevel Init函数的参数,用于设置log等级
func WithLevel(loglevel string) Option {
	return newFuncOption(func(o *Options) {
		o.Level = parseLevel(loglevel)
	})
}

//AddHooks Init函数的参数,用于增加钩子
func AddHooks(hooks ...logrus.Hook) Option {
	return newFuncOption(func(o *Options) {
		if o.Hooks == nil {
			o.Hooks = []logrus.Hook{}
		}
		o.Hooks = append(o.Hooks, hooks...)
	})
}

//WithDefaultFieldMap Init函数的参数,用于设置默认字段的新命名
func WithDefaultFieldMap(fm logrus.FieldMap) Option {
	return newFuncOption(func(o *Options) {
		o.DefaultFieldMap = fm
	})
}

//AddExtField Init函数的参数,用于增加扩展字段
func AddExtField(field string, value interface{}) Option {
	return newFuncOption(func(o *Options) {
		if o.ExtFields == nil {
			o.ExtFields = map[string]interface{}{field: value}
		} else {
			o.ExtFields[field] = value
		}
	})
}

//WithExtFields Init函数的参数,用于设置扩展字段
func WithExtFields(extFields map[string]interface{}) Option {
	return newFuncOption(func(o *Options) {
		if o.ExtFields == nil || len(o.ExtFields) == 0 {
			o.ExtFields = extFields
		} else {
			for field, value := range extFields {
				o.ExtFields[field] = value
			}
		}
	})
}

//WithOutput Init函数的参数,用于设置log的写入io
func WithOutput(writer io.Writer) Option {
	return newFuncOption(func(o *Options) {
		o.Output = writer
	})
}
