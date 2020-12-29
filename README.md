# loggerhelper

`github.com/sirupsen/logrus`的帮助程序.用于快速设置,开袋可用.

## 特性

+ 开袋可用,即便没有进行设置也已经设置为使用json格式打印消息.
+ WithFields可选,可以是0到任意多个
+ 可以通过`log.Init(loglevel string, defaultField map[string]interface{}, hooks ...logrus.Hook)`设置默认logger
+ 支持`hook`

## 基本用法

```golang

package main

import (
	log "github.com/Golang-Tools/loggerhelper"
)
func main() {
    log.Info("test")
    log.Warn("qweqwr", log.Dict{"a": 1})
}
```

## 初始化

```golang

package main

import (
	log "github.com/Golang-Tools/loggerhelper"
)
func main() {
    log.Init("WARN",log.Dict{"d":3})
    log.Info("test")
    log.Warn("qweqwr", log.Dict{"a": 1})
}
```