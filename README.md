# loggerhelper

`github.com/sirupsen/logrus`的帮助程序.用于快速设置,开袋可用.

## 特性

+ 开袋可用,即便没有进行设置也已经设置为使用json格式打印消息.
+ `WithExtFields`可选,可以是0到任意多个
+ 可以通过`log.Init(opts ...optparams.Option[log.Options])`设置默认logger
+ 支持`hook`

## 基本用法

即便不初始化也可以工作

```golang

package main

import (
    log "github.com/Golang-Tools/loggerhelper"
)
func main() {
    log.Info("test")
    log.Warn("qweqwr", log.Dict{"a": 1},log.Dict{"b": 1})
}
```

## 设置log

设置log可以设置log等级,默认字段的值,输出,以及钩子:

+ `func SetLogger(opts ...optparams.Option[Options])` 设置logger对象

```golang

package main

import (
    log "github.com/Golang-Tools/loggerhelper/v2"
    "io/ioutil"
    "os"
    "github.com/sirupsen/logrus"
    "github.com/sirupsen/logrus/hooks/writer"
)
func main() {
    log.Info("test1")
    log.SetLogger(log.WithLevel("WARN"), log.WithExtFields(log.Dict{"d": 3}))
    log.Info("test2")
    log.Warn("test3")
    log.SetLogger(log.WithLevel("Debug"), log.WithExtFields(log.Dict{"e": 3}))
    log.Debug("test4", log.Dict{"a": 1})
    log.Warn("test5", log.Dict{"a": 1})

    hook := writer.Hook{ // Send logs with level higher than warning to stderr
        Writer: os.Stderr,
        LogLevels: []logrus.Level{
            logrus.PanicLevel,
            logrus.FatalLevel,
            logrus.ErrorLevel,
            logrus.WarnLevel,
        },
    }
    log.SetLogger(WithLevel("WARN"), WithExtFields(log.Dict{"d": 3}), WithOutput(ioutil.Discard), AddHooks(hook))
    log.Info("test")
    log.Warn("qweqwr", log.Dict{"a": 1})
}
```

## 获取logger

获取logger接口`GetLogger() *logrus.Logger`可以获取到当前的logger,这主要用于导出给其他模块使用.
