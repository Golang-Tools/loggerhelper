package loggerhelper_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"

	log "github.com/Golang-Tools/loggerhelper/v4"
)

// resetGlobal 将全局logger复位到默认状态:json/debug/空扩展字段/无caller/无中间件
func resetGlobal(t *testing.T) {
	t.Helper()
	log.Set(
		log.WithOutput(io.Discard),
		log.WithLevel("Debug"),
		log.WithExtFields(map[string]interface{}{}),
		log.WithJSONFormat(),
		log.WithDisableReportCaller(),
		log.WithHandlerMiddlewares(),
	)
}

// resetToBuffer 复位全局状态并将日志输出到新的缓冲区,返回该缓冲区
func resetToBuffer(t *testing.T) *bytes.Buffer {
	t.Helper()
	resetGlobal(t)
	buf := &bytes.Buffer{}
	log.Set(log.WithOutput(buf), log.WithLevel("Debug"), log.WithExtFields(map[string]interface{}{}))
	return buf
}

// lastLine 解析缓冲区最后一条JSON日志为map
func lastLine(t *testing.T, buf *bytes.Buffer) map[string]interface{} {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &m); err != nil {
		t.Fatalf("解析日志失败: %v, 原文: %s", err, lines[len(lines)-1])
	}
	return m
}

func TestBasicJSONOutput(t *testing.T) {
	buf := resetToBuffer(t)
	log.Info("hello", log.Dict{"a": 1})
	m := lastLine(t, buf)
	if m["event"] != "hello" {
		t.Fatalf("event 字段错误: %v", m["event"])
	}
	if m["level"] != "info" {
		t.Fatalf("level 字段错误: %v", m["level"])
	}
	if m["a"] != float64(1) {
		t.Fatalf("附加字段 a 错误: %v", m["a"])
	}
	if _, ok := m["time"]; !ok {
		t.Fatal("缺少 time 字段")
	}
}

// TestExtFieldsNotBadKey 回归:扩展字段不得以 !BADKEY 输出
func TestExtFieldsNotBadKey(t *testing.T) {
	buf := resetToBuffer(t)
	log.Set(log.WithExtFields(log.Dict{"app": "l1"}))
	log.Info("hello")
	m := lastLine(t, buf)
	if m["app"] != "l1" {
		t.Fatalf("扩展字段 app 错误: %v", m)
	}
	if _, bad := m["!BADKEY"]; bad {
		t.Fatalf("出现了 !BADKEY: %v", m)
	}
}

func TestLevelFilter(t *testing.T) {
	buf := resetToBuffer(t)
	log.Set(log.WithLevel("Warn"))
	buf.Reset()
	log.Info("filtered")
	if strings.Contains(buf.String(), "filtered") {
		t.Fatal("Warn 阈值下不应输出 Info")
	}
	log.Warn("kept")
	if !strings.Contains(buf.String(), "kept") {
		t.Fatal("Warn 阈值下应输出 Warn")
	}
	resetGlobal(t)
}

func TestTextFormatAndSwitchBack(t *testing.T) {
	buf := &bytes.Buffer{}
	log.Set(log.WithOutput(buf), log.WithLevel("Debug"), log.WithTextFormat(), log.WithExtFields(map[string]interface{}{}))
	log.Info("plain")
	if !strings.Contains(buf.String(), "level=info") {
		t.Fatalf("text 格式缺少 level=info: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "event=plain") {
		t.Fatalf("text 格式缺少 event 字段: %s", buf.String())
	}
	// 验证可切回 json
	resetGlobal(t)
	buf2 := &bytes.Buffer{}
	log.Set(log.WithOutput(buf2))
	log.Info("back to json")
	if !strings.Contains(buf2.String(), `"event":"back to json"`) {
		t.Fatalf("切回 json 失败: %s", buf2.String())
	}
	resetGlobal(t)
}

func TestExportKeepsFieldsAndNeverNil(t *testing.T) {
	buf := resetToBuffer(t)
	log.Set(log.WithExtFields(log.Dict{"app": "l1"}))
	lg := log.Export()
	if lg == nil {
		t.Fatal("Export 返回了 nil")
	}
	lg.Debug("from export")
	m := lastLine(t, buf)
	if m["event"] != "from export" {
		t.Fatalf("导出的 Log 未正常工作: %v", m)
	}
	if m["app"] != "l1" {
		t.Fatalf("导出对象应固化导出时的字段: %v", m)
	}
	resetGlobal(t)
}

func TestExportFollowsLaterSetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	log.Set(log.WithOutput(buf), log.WithLevel("Debug"), log.WithExtFields(log.Dict{"app": "l1"}))
	lg := log.Export()
	// 后续 Set 提高等级,导出的 Log 应跟随(等级经共享levelVar即时生效)
	log.Set(log.WithLevel("Warn"), log.WithExtFields(log.Dict{"app": "l1"}))
	buf.Reset()
	lg.Info("filtered")
	if strings.Contains(buf.String(), "filtered") {
		t.Fatal("导出的 Log 应跟随后续 Set 的等级,Warn 下不应输出 Info")
	}
	lg.Warn("passes")
	m := lastLine(t, buf)
	if m["event"] != "passes" {
		t.Fatalf("导出的 Log 未跟随后续 Set: %v", m)
	}
	if m["app"] != "l1" {
		t.Fatalf("导出对象应保留导出时的固化字段,实际: %v", m["app"])
	}
	resetGlobal(t)
}

func TestReportCallerPointsToCallSite(t *testing.T) {
	buf := &bytes.Buffer{}
	log.Set(log.WithOutput(buf), log.WithLevel("Debug"), log.WithReportCaller(), log.WithExtFields(map[string]interface{}{}))
	log.Info("with caller")
	m := lastLine(t, buf)
	file, ok := m["file"].(string)
	if !ok {
		t.Fatalf("未输出 file 字段: %v", m)
	}
	if !strings.Contains(file, "loggerhelper_test.go") {
		t.Fatalf("caller 未指向调用点,而是: %s", file)
	}
	// 关闭 caller 后不再输出
	resetGlobal(t)
	buf2 := &bytes.Buffer{}
	log.Set(log.WithOutput(buf2), log.WithLevel("Debug"))
	log.Info("no caller")
	m2 := lastLine(t, buf2)
	if _, ok := m2["file"]; ok {
		t.Fatalf("关闭 caller 后不应输出 file: %v", m2)
	}
	resetGlobal(t)
}

func TestPanicLevelPanics(t *testing.T) {
	log.Set(log.WithOutput(io.Discard), log.WithLevel("Debug"), log.WithExtFields(map[string]interface{}{}))
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Panic 未触发 panic")
		}
		resetGlobal(t)
	}()
	log.Panic("boom")
}

func TestTraceOnlyWhenEnabled(t *testing.T) {
	buf := resetToBuffer(t)
	log.Trace("t1") // 默认 Debug 阈值,Trace 应被过滤
	if strings.Contains(buf.String(), "t1") {
		t.Fatal("Debug 阈值下不应输出 Trace")
	}
	log.Set(log.WithLevel("Trace"))
	buf.Reset()
	log.Trace("t2")
	if !strings.Contains(buf.String(), "t2") {
		t.Fatal("Trace 阈值下应输出 Trace")
	}
	resetGlobal(t)
}

func TestFanoutOutput(t *testing.T) {
	resetGlobal(t)
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	h2 := slog.NewJSONHandler(buf2, nil)
	log.Set(log.WithOutput(buf1), log.WithLevel("Debug"), log.WithFanoutOutput(h2))
	log.Info("fan")
	if !strings.Contains(buf1.String(), "fan") {
		t.Fatalf("主输出未收到日志: %s", buf1.String())
	}
	if !strings.Contains(buf2.String(), "fan") {
		t.Fatalf("扇出输出未收到日志: %s", buf2.String())
	}
	resetGlobal(t)
}

func TestLevelRoute(t *testing.T) {
	resetGlobal(t)
	low := &bytes.Buffer{}
	err := &bytes.Buffer{}
	lerr := slog.NewJSONHandler(err, nil)
	log.Set(log.WithOutput(low), log.WithLevel("Debug"), log.WithLevelRoute(slog.LevelError, lerr))
	log.Info("low level")
	log.Error("high level")
	if !strings.Contains(low.String(), "low level") {
		t.Fatalf("低等级应走主输出: %s", low.String())
	}
	if strings.Contains(low.String(), "high level") {
		t.Fatalf("Error 不应走主输出: %s", low.String())
	}
	if !strings.Contains(err.String(), "high level") {
		t.Fatalf("Error 应走路由输出: %s", err.String())
	}
	resetGlobal(t)
}

func TestConcurrentSetAndLog(t *testing.T) {
	resetGlobal(t)
	log.Set(log.WithOutput(io.Discard), log.WithLevel("Debug"), log.WithExtFields(map[string]interface{}{}))
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				log.Info("concurrent", log.Dict{"j": j})
			}
		}()
	}
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				log.Set(log.WithLevel("Debug"), log.WithExtFields(log.Dict{"app": "x"}))
			}
		}()
	}
	wg.Wait()
	resetGlobal(t)
}

func TestConcurrentSetAndExportedLog(t *testing.T) {
	resetGlobal(t)
	log.Set(log.WithOutput(io.Discard), log.WithLevel("Debug"), log.WithExtFields(map[string]interface{}{}))
	lg := log.Export()
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				lg.Info("exported concurrent", log.Dict{"j": j})
			}
		}()
	}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				log.Set(log.WithLevel("Debug"), log.WithExtFields(log.Dict{"app": "x"}))
			}
		}()
	}
	wg.Wait()
	resetGlobal(t)
}

func TestUnknownLevelKeepsPrevious(t *testing.T) {
	buf := resetToBuffer(t)
	log.Set(log.WithLevel("Warn"))
	log.Set(log.WithLevel("not-a-level"))
	buf.Reset()
	log.Info("should be filtered")
	if strings.Contains(buf.String(), "should be filtered") {
		t.Fatal("非法等级不应改变阈值")
	}
	resetGlobal(t)
}
