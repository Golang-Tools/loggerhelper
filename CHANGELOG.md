# 2.0.0

V2版本是对V0版本的重构,允许修改全局logger,并允许将logger输出

V2版本针对go 1.18+,使用泛型语法.低版本还是继续使用v0版本

# 0.0.4

## 修改接口

+ 使用`Opt`的形式统一Init接口
+ 增加对`TextFormat`的支持
+ 增加对设置时间戳字符串的格式的支持
+ 增加对是否输出时间的选项
+ 增加对修改默认字段(time,level,event,logrus_error,caller,file)字段名的支持

## 依赖更新

更新到`github.com/sirupsen/logrus@v1.8.1`

# 0.0.3

## 修改接口

+ 拆分初始化函数

## 修复bug

+ 初始化后第一个filed不会打印的bug

# 0.0.2

## 新增功能

+ 增加了添加hook的能力
+ 增加了设置output的能力

## 修改接口

+ 修改log接口,更加符合一般用法

## 修改依赖

+ "github.com/sirupsen/logrus v1.6.0" -> "github.com/sirupsen/logrus v1.7.0"