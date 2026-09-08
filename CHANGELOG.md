# 4.0.0

V4版本使用 Go 标准库 `log/slog` 重写,不再依赖 `github.com/sirupsen/logrus`,仅依赖 `github.com/Golang-Tools/optparams`。面向 go 1.22+。

## 破坏性变更

+ 模块路径变更为 `github.com/Golang-Tools/loggerhelper/v4`
+ 最低 go 版本提升到 1.22
+ 移除 logrus 依赖与所有 logrus 类型;需要 logrus 实现请继续使用 `/v3`(`v3.0.0` 起)或 `/v2`
+ 接口等价替换:
  + `GetLogger()` 返回值由 `*logrus.Logger` / `*logrus.Entry` 改为 `*slog.Logger`
  + `AddHooks(...logrus.Hook)` 由 `WithHandlerMiddleware(...)`(追加)/ `WithHandlerMiddlewares(...)`(整体设置)替代
  + `WithDefaultFieldMap(logrus.FieldMap)` 由内置 v3 兼容风格 + `WithReplaceAttr(fn)` 替代
  + `Options.Level` / `Options.Hooks` 等字段改为 slog 对应类型
+ `New()` 返回独立配置的 `*slog.Logger`

## 新增接口

+ 自定义级别:`LevelTrace`(debug-4)、`LevelFatal`(error+4)、`LevelPanic`(error+8)
+ `WithHandlerMiddleware(mws...)` 追加 handler 中间件;`WithHandlerMiddlewares(mws...)` 整体设置(可清空)
+ 内置中间件:`WithFanoutOutput(handlers...)` 多路输出;`WithLevelRoute(threshold, handler)` 按级别路由
+ `WithReplaceAttr(fn)` 完全自定义字段改写
+ `WithJSONFormat()` / `WithTextFormat()` 双向切换输出格式
+ `WithDisableReportCaller()` 关闭调用方信息
+ `Log.GetLogger() *slog.Logger` 获取携带固化字段的 logger

## 行为说明

+ 字段风格与 v3 兼容:`event` 表示消息、`level` 输出小写文本(warn 为 warning)、时间默认为 RFC3339Nano
+ 等级通过共享的 `LevelVar` 即时生效:即使已通过 `GetLogger()`/`Export()` 取走的 logger,其等级阈值仍跟随后续 `log.Set`
+ `Set()` 为增量配置,基于上一次选项累积

## 维护约定

+ v4 为纯 `log/slog` 实现,**不得引入 logrus 依赖**。每次改动后请运行 `scripts/check_no_logrus.sh` 校验(等价 `go list -deps ./... | grep logrus` 无输出)。

# 3.0.0

V3版本在保持V2面向应用接口不变的前提下,对基于logrus的实现做现代化与缺陷修复,面向go 1.22+。(以独立分支 `/v3` 维护,标签 `v3.0.0`)

## 破坏性变更

+ 模块路径变更为 `github.com/Golang-Tools/loggerhelper/v3`
+ 最低 go 版本提升到 1.22,依赖 `github.com/Golang-Tools/optparams` 升级到 v1.0.0
+ 包级 `GetLogger()` 返回调用时刻的 logger 快照;若在其之后调用 `Set()`,该引用不再反映新配置。需要跟随后续动态调整时,请使用包级日志函数或 `Export()` 得到的 `Log` 对象,这两者不受影响

## bug 修复

+ 修复并发调用 `Set()` 与打印日志之间的数据竞争(`go test -race` 通过)
+ 修复重复调用 `Set()` 时 hook 不断累加的问题,改为每次整体重建 hooks
+ 修复配置了 `caller`/`file` 字段却不生效的问题,新增 `WithReportCaller()` 开启后调用方信息会正确指向业务代码而非本包
+ 修复 `AddExtField`/`WithAddExtFields` 就地修改共享 map 导致污染包级 `DefaultOpts` 的问题
+ 修复 `Export()` 在无扩展字段时返回 `nil` 导致调用方 panic 的问题,现在恒返回可用对象
+ 未知日志等级不再通过 `fmt.Printf` 打印到标准输出,改为忽略并保持原有等级

## 新增接口

+ 新增 `WithReportCaller()` 选项用于开启调用方信息输出

# 2.0.2

## 新增接口

+ 增加Log对象的`GetLogger`接口用于获取`logrus.Entry`对象

# 2.0.1

## bug修复

+ 修复`ExtFields`置空时的行为,`ExtFields`为空字典时`defaultlog`会被置为`nil`

# 2.0.0

V2版本是对V0版本的重构,允许修改全局logger,并允许将logger输出

V2版本针对go 1.18+,使用泛型语法.低版本还是继续使用v0版本

主要改变为:

+ 取消`Init()`接口,改为`Set()`接口,现在`Set()`接口用于修改logger设置.
+ 取消外部变量`Logger`,改为使用`GetLogger()`接口获取对象
+ 取消模块中直接赋值变量,使用`init()`方式初始化模块
+ 新增`Log`类型,用于固定有特定固定ExtFields的log,使用`Export()`导出当前ExtFields的log对象.
