package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"

	log "github.com/Golang-Tools/loggerhelper/v4"
)

// newBuf 复位全局logger,使其写入一个新的缓冲区,便于每段演示互不影响
func newBuf() *bytes.Buffer {
	buf := &bytes.Buffer{}
	log.Set(
		log.WithOutput(buf),
		log.WithLevel("Debug"),
		log.WithJSONFormat(),
		log.WithExtFields(log.Dict{}),
		log.WithDisableReportCaller(),
		log.WithHandlerMiddlewares(),
	)
	return buf
}

// show 打印一段演示的标题与输出
func show(title string, buf *bytes.Buffer) {
	fmt.Printf("=== %s ===\n%s\n", title, buf.String())
}

func demoBasic() {
	buf := newBuf()
	log.Info("test")
	log.Warn("qweqwr", log.Dict{"a": 1}, log.Dict{"b": 1})
	show("开袋即用:不初始化直接输出JSON(字段可合并多个Dict)", buf)
}

func demoLevelFilter() {
	buf := newBuf()
	log.Set(log.WithLevel("WARN"))
	log.Info("这条 Info 会被过滤")
	log.Warn("这条 Warn 会输出")
	show("等级过滤(WithLevel WARN)", buf)
}

func demoExtFields() {
	buf := newBuf()
	log.Set(log.WithExtFields(log.Dict{"app": "l1"}))
	log.Info("带扩展字段")
	log.Set(log.WithAddExtFields(log.Dict{"region": "cn"}))
	log.Info("追加扩展字段")
	log.Set(log.AddExtField("env", "prod"))
	log.Info("再追加单个字段")
	show("扩展字段(WithExtFields / WithAddExtFields / AddExtField)", buf)
}

func demoExport() {
	buf := newBuf()
	log.Set(log.WithExtFields(log.Dict{"app": "l1"}))
	logger1 := log.Export()
	log.Set(log.WithExtFields(log.Dict{"app": "l2"}))
	logger2 := log.Export()
	logger1.Debug("logger1 的日志")
	logger2.Debug("logger2 的日志")
	show("导出Log:不同业务不同标识(Export),字段固化、等级/output跟随全局", buf)
}

func demoCaller() {
	buf := newBuf()
	log.Set(log.WithReportCaller())
	log.Info("带调用方信息")
	show("调用方信息(WithReportCaller),指向本函数内的调用点", buf)
}

func demoTextFormat() {
	buf := newBuf()
	log.Set(log.WithTextFormat())
	log.Info("text 风格")
	log.Set(log.WithJSONFormat())
	log.Info("再切回 json")
	show("输出格式切换(WithTextFormat / WithJSONFormat)", buf)
}

func demoFanout() {
	buf := newBuf()
	extra := &bytes.Buffer{}
	log.Set(log.WithFanoutOutput(slog.NewTextHandler(extra, nil)))
	log.Info("同时发往两个输出")
	fmt.Printf("=== fan-out:主输出(JSON)+额外handler(Text) ===\n主输出: %s额外输出: %s\n", buf.String(), extra.String())
}

func demoLevelRoute() {
	buf := newBuf()
	errBuf := &bytes.Buffer{}
	log.Set(log.WithLevelRoute(slog.LevelError, slog.NewJSONHandler(errBuf, nil)))
	log.Info("低等级走主输出")
	log.Error("error及以上走路由输出")
	fmt.Printf("=== 按级别路由(Error+ 交给独立handler) ===\n主输出: %s路由输出: %s\n", buf.String(), errBuf.String())
}

func demoGetLogger() {
	buf := newBuf()
	raw := log.GetLogger() // *slog.Logger,可直接交给接受 slog.Logger 的第三方组件
	raw.Info("来自 GetLogger 的 slog.Logger", "key", "value")
	show("获取logger(GetLogger → *slog.Logger)", buf)
}

func main() {
	demoBasic()
	demoLevelFilter()
	demoExtFields()
	demoExport()
	demoCaller()
	demoTextFormat()
	demoFanout()
	demoLevelRoute()
	demoGetLogger()

	// 收尾:回到默认 stderr 输出
	log.Set(log.WithOutput(os.Stderr), log.WithLevel("Debug"), log.WithJSONFormat(), log.WithExtFields(log.Dict{}), log.WithDisableReportCaller(), log.WithHandlerMiddlewares())
	log.Info("示例结束(以下回到默认 stderr 输出)")
}
