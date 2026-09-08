# loggerhelper/V3

`github.com/sirupsen/logrus`的帮助程序.用于快速设置,开袋可用.

该模块定位为基础组件,如果需要用它就只用它比较好.

V3版本在保持V2面向应用接口不变的前提下,对基于logrus的实现做现代化与缺陷修复(并发安全/hook重建/caller生效等).

V3版本针对go 1.22+,依赖`optparams` v1.0.0.低版本请继续使用v2或v0版本

## 特性

+ 开袋可用,即便没有进行设置也已经设置为使用json格式打印消息.
+ `WithExtFields`可选,可以是0到任意多个
+ 可以通过`log.Set(opts ...optparams.Option[log.Options])`设置默认logger
+ 支持`hook`
+ 支持获取默认logger
+ 支持导出Log对象针对不同业务进行区分.导出的Log对象只是默认字段不同,依然受log.Set设置的其他属性影响

## 基本用法

即便不初始化也可以工作

```golang

package main

import (
    log "github.com/Golang-Tools/loggerhelper/v3"
)
func main() {
    log.Info("test")
    log.Warn("qweqwr", log.Dict{"a": 1},log.Dict{"b": 1})
}
```

## 设置log

设置log可以设置log等级,默认字段的值,输出,以及钩子:

+ `func Set(opts ...optparams.Option[Options])` 设置logger对象

```golang

package main

import (
    log "github.com/Golang-Tools/loggerhelper/v3"
    "io/ioutil"
    "os"
    "github.com/sirupsen/logrus"
    "github.com/sirupsen/logrus/hooks/writer"
)
func main() {
    log.Info("test1")
    log.Set(log.WithLevel("WARN"), log.WithExtFields(log.Dict{"d": 3}))
    log.Info("test2")
    log.Warn("test3")
    log.Set(log.WithLevel("Debug"), log.WithExtFields(log.Dict{"e": 3}))
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
    log.Set(log.WithLevel("WARN"), log.WithExtFields(log.Dict{"d": 3}), log.WithOutput(ioutil.Discard), log.AddHooks(hook))
    log.Info("test")
    log.Warn("qweqwr", log.Dict{"a": 1})
}
```

## 输出调用方信息

默认不输出调用方信息.通过`WithReportCaller()`可以开启,开启后日志会带上`caller`(函数名)与`file`(文件:行号)字段,且会正确指向业务调用处而非本模块内部:

```go
log.Set(log.WithReportCaller())
log.Info("with caller")
// {"caller":"main.main","event":"with caller","file":".../main.go:12","level":"info",...}
```

## 获取logger

获取logger接口`GetLogger() *logrus.Logger`可以获取到当前的logger,这主要用于导出给其他模块使用.比如用于设置gin的log

注意:`GetLogger()`返回的是调用时刻的logger快照,之后调用`log.Set()`不会改变已取走的这个对象.需要跟随后续动态调整时请使用包级日志函数,或`Export()`得到的`Log`对象(其等级/output等仍受后续`log.Set`影响)

```go
app.Use(ginlogrus.Logger(log.GetLogger()), gin.Recovery())
```

## 导出Log

往往我们希望不同的模块使用不同的log标识这样可以比较准确地定位错误.这时可以通过`Export`接口导出当前的Log对象.

```go
log.Set(log.WithExtFields(log.Dict{"app": "l1"}))
Logger1 := log.Export()

log.Set(log.WithExtFields(log.Dict{"app": "l2"}))
Logger2 := log.Export()
Logger1.Debug("test logger1")
Logger2.Debug("test logger2")
```
