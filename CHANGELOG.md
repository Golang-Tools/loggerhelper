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
