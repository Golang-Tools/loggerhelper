# loggerhelper/V4

基于 Go 标准库 `log/slog` 的结构化日志帮助程序,用于快速设置,开袋可用。

该模块定位为基础组件,如果需要用它就只用它比较好。

V4版本使用标准库 `log/slog` 重写,**不依赖 logrus**(仅依赖 `github.com/Golang-Tools/optparams`),面向 go 1.22+。需要继续使用 logrus 的实现请使用 `/v3`(见 [v3 分支](https://github.com/Golang-Tools/loggerhelper/tree/v3))或 `/v2`。

## 如何选择版本

| 版本 | 底层 | 适用场景 | 引入路径 |
| --- | --- | --- | --- |
| **v4(当前,推荐)** | 标准库 `log/slog`,仅依赖 `optparams` | 新项目首选;要求 go 1.22+ | `github.com/Golang-Tools/loggerhelper/v4`(master) |
| v3 | logrus(现代化修复,go 1.22+) | 存量项目仍重度依赖 logrus 的 hook 生态 | `github.com/Golang-Tools/loggerhelper/v3`(v3 分支) |
| v2 | logrus(旧,go 1.18+) | 最低 go 版本兼容 | `github.com/Golang-Tools/loggerhelper/v2`(历史 tag) |

> 面向应用的高层接口(日志函数、`Set`、`Export`、各 `With*` 选项)在 v2/v3/v4 之间保持一致。只要你不依赖 logrus 的 hook,**新项目直接引 v4、存量项目通常只改 import 路径即可**;需要改动的地方见文末"迁移自 v2 / v3"。

## 特性

+ 开袋可用,默认即使用 json 格式打印消息
+ 字段风格与 v3 保持兼容:`event` 表示消息、`level` 为小写文本、`time` 时间戳
+ 通过 `log.Set(opts ...optparams.Option[Options])` 增量配置全局 logger
+ 等级动态生效:已取走的 logger/导出的 Log 其等级阈值仍跟随后续 `log.Set`
+ 支持 handler 中间件(替代 logrus 的 hook),内置多路输出与按级别路由
+ 支持获取默认 logger
+ 支持导出 Log 对象针对不同业务进行区分,导出的 Log 对象只固化默认字段,等级/output 等依然受 `log.Set` 设置影响

## 基本用法

即便不初始化也可以工作

```golang

package main

import (
    log "github.com/Golang-Tools/loggerhelper/v4"
)
func main() {
    log.Info("test")
    log.Warn("qweqwr", log.Dict{"a": 1}, log.Dict{"b": 1})
}
// {"time":"...","level":"info","event":"test"}
// {"time":"...","level":"warning","event":"qweqwr","a":1,"b":1}
```

## 设置log

`Set(opts ...optparams.Option[Options])` 为增量配置,会累积到全局选项上。可用选项:

+ `WithLevel("debug"|"info"|"warn"|"error"|"trace"|"fatal"|"panic")` 设置等级(大小写不敏感),未知值被忽略
+ `WithExtFields(map)` 重置扩展字段;`AddExtField(k,v)` / `WithAddExtFields(map)` 追加
+ `WithOutput(io.Writer)` 设置输出(默认 stderr)
+ `WithTextFormat()` / `WithJSONFormat()` 切换输出格式(默认 json)
+ `WithDisableTimeField()` 关闭时间字段;`WithTimeFormat(layout)` 设置时间格式
+ `WithReportCaller()` / `WithDisableReportCaller()` 开关调用方信息
+ `WithReplaceAttr(fn)` 自定义字段改写(完全接管内置的 v3 兼容风格)
+ `WithHandlerMiddleware(mw)` / `WithHandlerMiddlewares(...)` 增加/整体设置 handler 中间件

```golang

package main

import (
    log "github.com/Golang-Tools/loggerhelper/v4"
    "io"
)
func main() {
    log.Info("test1")
    log.Set(log.WithLevel("WARN"), log.WithExtFields(log.Dict{"d": 3}))
    log.Info("test2") // 被过滤
    log.Warn("test3")
    log.Set(log.WithLevel("Debug"), log.WithExtFields(log.Dict{"e": 3}))
    log.Debug("test4", log.Dict{"a": 1})
    log.Warn("test5", log.Dict{"a": 1})

    log.Set(log.WithLevel("WARN"), log.WithExtFields(log.Dict{"d": 3}), log.WithOutput(io.Discard))
    log.Info("test")
    log.Warn("qweqwr", log.Dict{"a": 1})
}
```

## handler 中间件(替代 hook)

V4 用 handler 中间件替代 logrus 的 hook。中间件包装 `slog.Handler` 以扩展/分流日志。

### 多路输出(fan-out)

```golang
import (
    "log/slog"
    log "github.com/Golang-Tools/loggerhelper/v4"
)

func main() {
    // 除主输出外,同时把日志发给一个 text 格式 handler
    textH := slog.NewTextHandler(os.Stdout, nil)
    log.Set(log.WithFanoutOutput(textH))
    log.Info("both outputs")
}
```

### 按级别路由

```golang
func main() {
    errH := slog.NewJSONHandler(errFile, nil)
    // error 及以上单独交给 errH,其余走主输出
    log.Set(log.WithLevelRoute(slog.LevelError, errH))
    log.Info("main out")
    log.Error("err out")
}
```

### 自定义中间件

```golang
func main() {
    log.Set(log.WithHandlerMiddleware(func(next slog.Handler) slog.Handler {
        return myHandler{next: next}
    }))
}
```

## 调用方信息

`WithReportCaller()` 开启后,日志会带上 `caller`(函数名)与 `file`(文件:行号),并指向业务调用处:

```golang
log.Set(log.WithReportCaller())
log.Info("with caller")
// {"time":"...","level":"info","event":"with caller","caller":"main.main","file":".../main.go:12"}
```

## 获取logger

`GetLogger() *slog.Logger` 返回当前的基础 logger(不含扩展字段),主要用于导出给其他模块使用,如接入 HTTP 框架:

```go
srv.Handler = myMiddleware(log.GetLogger())
```

注意:返回的 logger 其**等级阈值**仍跟随后续 `log.Set`(通过共享的 LevelVar 即时生效),但输出目标/格式等在取走那一刻固定。需要完整跟随动态配置请使用包级函数或 `Export()` 得到的 `Log` 对象。

## 导出Log

通过 `Export()` 导出固化当前扩展字段的 `Log` 对象,便于不同模块使用不同标识:

```go
log.Set(log.WithExtFields(log.Dict{"app": "l1"}))
Logger1 := log.Export()

log.Set(log.WithExtFields(log.Dict{"app": "l2"}))
Logger2 := log.Export()
Logger1.Debug("test logger1") // app=l1
Logger2.Debug("test logger2") // app=l2
```

## 自定义字段风格

默认字段风格与 v3 兼容:`event`(消息)、小写 `level`、RFC3339Nano 的 `time`。需要完全自定义时可使用 `WithReplaceAttr`:

```go
log.Set(log.WithReplaceAttr(func(groups []string, a slog.Attr) slog.Attr {
    if a.Key == slog.MessageKey {
        return slog.String("msg", a.Value.String())
    }
    return a
}))
```

## 迁移自 v2 / v3

面向应用的高层接口(日志函数、`Set`、`Export`、选项)签名基本不变,改动集中在底层泄漏类型:

| v2/v3(logrus) | v4(slog) |
| --- | --- |
| `GetLogger() *logrus.Logger` / `*logrus.Entry` | `GetLogger() *slog.Logger` |
| `AddHooks(...logrus.Hook)` | `WithHandlerMiddleware` / `WithFanoutOutput` / `WithLevelRoute` |
| `WithDefaultFieldMap(logrus.FieldMap)` | 默认内置 v3 兼容风格,或用 `WithReplaceAttr` 自定义 |
| `Options.Level` (logrus.Level) | `Options.Level` (slog.Level) |
| `New() *logrus.Logger` | `New() *slog.Logger` |

### 示例:基础用法几乎无需改动

```go
// v2/v3(logrus)
import log "github.com/Golang-Tools/loggerhelper/v3"
log.Set(log.WithLevel("WARN"), log.WithExtFields(log.Dict{"app": "svc"}))
log.Warn("oops", log.Dict{"id": 1})

// v4(slog):仅换 import 路径,API 一致
import log "github.com/Golang-Tools/loggerhelper/v4"
log.Set(log.WithLevel("WARN"), log.WithExtFields(log.Dict{"app": "svc"}))
log.Warn("oops", log.Dict{"id": 1})
// 两者输出均为 {"time":...,"level":"warning","event":"oops","app":"svc","id":1}
```

### 示例:hook → handler 中间件

logrus 的 hook 常见用法是"把 warn 及更严重的日志额外写一份到 stderr",v4 中可用内置的 fan-out + 带等级的 handler 等价实现:

```go
// v2/v3(logrus):writer.Hook 把 warn 及以上额外写一份到 stderr
hook := &writer.Hook{Writer: os.Stderr, LogLevels: []logrus.Level{
    logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel, logrus.WarnLevel,
}}
log.Set(log.AddHooks(hook))

// v4(slog):highH 自带 warn 阈值,fan-out 让它与主输出并存
highH := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn})
log.Set(log.WithFanoutOutput(highH))
```

若想"分流替换"(warn 及以上不再进主输出、单独交给某 handler),则用按级别路由:

```go
highH := slog.NewJSONHandler(os.Stderr, nil)
log.Set(log.WithLevelRoute(slog.LevelWarn, highH))
```

### 示例:GetLogger 返回类型

`GetLogger()` 由 `*logrus.Logger` 变为 `*slog.Logger`,用于接入接受 `*slog.Logger` 的第三方组件;若对方只接受 logrus,可在 v4 侧做一层薄适配,或继续使用 v3。

```go
// v4:返回 *slog.Logger
l := log.GetLogger() // *slog.Logger
```
