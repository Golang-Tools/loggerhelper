# 3.0.0

V3 版本在保持 V2 面向应用接口不变的前提下,对基于 logrus 的实现做现代化与缺陷修复,面向 go 1.22+。

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
