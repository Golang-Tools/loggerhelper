# loggerhelper

`github.com/sirupsen/logrus`的帮助程序.用于快速设置,开袋可用.

## 特性

+ 开袋可用,即便没有进行设置也已经设置为使用json格式打印消息.
+ `WithExtFields`可选,可以是0到任意多个
+ 可以通过`log.Init(opts ...Option)`设置默认logger
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

## 初始化

初始化可以设置log等级,默认字段的值,输出,以及钩子,初始化有两个方法:

+ `Init(WithLevel(loglevel),...)` 初始化默认的log对象

```golang

package main

import (
    log "github.com/Golang-Tools/loggerhelper"
    "io/ioutil"
    "os"
    "github.com/sirupsen/logrus"
    "github.com/sirupsen/logrus/hooks/writer"
)
func main() {
    hook := writer.Hook{ // Send logs with level higher than warning to stderr
        Writer: os.Stderr,
        LogLevels: []logrus.Level{
            logrus.PanicLevel,
            logrus.FatalLevel,
            logrus.ErrorLevel,
            logrus.WarnLevel,
        },
    }
    log.Init(WithLevel("WARN"), WithExtFields(log.Dict{"d": 3}), WithOutput(ioutil.Discard), AddHooks(hook))
    log.Info("test")
    log.Warn("qweqwr", log.Dict{"a": 1})
}
```
