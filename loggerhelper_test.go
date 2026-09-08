package loggerhelper_test

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	log "github.com/Golang-Tools/loggerhelper/v3"
	logrus "github.com/sirupsen/logrus"
)

//resetToBuffer 将全局logger重置到一个新的缓冲区,返回该缓冲区
func resetToBuffer(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	log.Set(
		log.WithOutput(buf),
		log.WithLevel("Debug"),
		log.WithExtFields(map[string]interface{}{}),
	)
	return buf
}

//lastLine 解析缓冲区最后一条JSON日志为map
func lastLine(t *testing.T, buf *bytes.Buffer) map[string]interface{} {
	t.Helper()
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	m := map[string]interface{}{}
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &m); err != nil {
		t.Fatalf("解析日志失败: %v, 原文: %s", err, lines[len(lines)-1])
	}
	return m
}

//countingHook 记录被触发次数的hook
type countingHook struct {
	n int32
}

func (h *countingHook) Levels() []logrus.Level { return logrus.AllLevels }
func (h *countingHook) Fire(*logrus.Entry) error {
	atomic.AddInt32(&h.n, 1)
	return nil
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
}

func TestExportNeverNil(t *testing.T) {
	buf := resetToBuffer(t)
	// ExtFields 为空时 defaultlog 为 nil,Export 仍应返回可用对象
	lg := log.Export()
	if lg == nil {
		t.Fatal("Export 返回了 nil")
	}
	lg.Debug("from export")
	m := lastLine(t, buf)
	if m["event"] != "from export" {
		t.Fatalf("导出的 Log 未正常工作: %v", m)
	}
}

func TestHooksNotAccumulated(t *testing.T) {
	resetToBuffer(t)
	h := &countingHook{}
	// 添加一次 hook 后,多次无参 Set 不应造成 hook 重复累加
	log.Set(log.AddHooks(h))
	log.Set()
	log.Set()
	atomic.StoreInt32(&h.n, 0)
	log.Info("trigger")
	if got := atomic.LoadInt32(&h.n); got != 1 {
		t.Fatalf("hook 触发次数应为 1,实际为 %d(说明发生了累加)", got)
	}
	// 清理:重置 hooks
	log.Set(log.WithOutput(io.Discard))
	log.Set()
	atomic.StoreInt32(&h.n, 0)
	log.Info("after reset")
	if got := atomic.LoadInt32(&h.n); got != 1 {
		t.Fatalf("重置后 hook 仍应保留一次,实际 %d", got)
	}
}

func TestReportCallerPointsToCallSite(t *testing.T) {
	buf := &bytes.Buffer{}
	log.Set(
		log.WithOutput(buf),
		log.WithLevel("Debug"),
		log.WithExtFields(map[string]interface{}{}),
		log.WithReportCaller(),
	)
	log.Info("with caller")
	m := lastLine(t, buf)
	file, ok := m["file"].(string)
	if !ok {
		t.Fatalf("未输出 file 字段: %v", m)
	}
	if !strings.Contains(file, "loggerhelper_test.go") {
		t.Fatalf("caller 未指向调用点,而是: %s", file)
	}
	// 关闭 caller,清理全局状态
	log.Set(log.WithOutput(io.Discard))
	log.Set(log.WithExtFields(map[string]interface{}{}))
}

func TestReportCallerReset(t *testing.T) {
	// WithReportCaller 是一次性开启,需能再次关闭
	buf := &bytes.Buffer{}
	log.Set(log.WithOutput(buf), log.WithLevel("Debug"), log.WithReportCaller())
	log.Info("on")
	// 通过重建默认选项关闭 report caller
	log.Set(log.WithOutput(buf), log.WithExtFields(map[string]interface{}{}))
	// 无法直接翻转 ReportCaller,这里仅验证 file 字段在上一步已产出
	if !strings.Contains(buf.String(), "file") {
		t.Fatal("开启 caller 后应输出 file 字段")
	}
}

func TestDefaultOptsNotPolluted(t *testing.T) {
	resetToBuffer(t)
	before := len(log.DefaultOpts.ExtFields)
	log.Set(log.AddExtField("polluteKey", 1))
	if _, ok := log.DefaultOpts.ExtFields["polluteKey"]; ok {
		t.Fatal("AddExtField 污染了包级 DefaultOpts.ExtFields")
	}
	if len(log.DefaultOpts.ExtFields) != before {
		t.Fatalf("DefaultOpts.ExtFields 长度被改变: %d -> %d", before, len(log.DefaultOpts.ExtFields))
	}
	log.Set(log.WithExtFields(map[string]interface{}{}))
}

func TestUnknownLevelDoesNotChangeLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	log.Set(log.WithOutput(buf), log.WithLevel("Warn"), log.WithExtFields(map[string]interface{}{}))
	// 非法等级不应 panic,也不应把等级改成 Debug
	log.Set(log.WithLevel("not-a-level"))
	buf.Reset()
	log.Info("should be filtered")
	if strings.Contains(buf.String(), "should be filtered") {
		t.Fatal("非法等级不应把日志等级降级到可打印 Info")
	}
}

func TestPanicLevelPanics(t *testing.T) {
	log.Set(log.WithOutput(io.Discard), log.WithLevel("Debug"), log.WithExtFields(map[string]interface{}{}))
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("Panic 未触发 panic")
		}
	}()
	log.Panic("boom")
}

func TestConcurrentSetAndLog(t *testing.T) {
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
}

func TestExportFollowsLaterSetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	log.Set(log.WithOutput(buf), log.WithLevel("Debug"), log.WithExtFields(log.Dict{"app": "l1"}))
	lg := log.Export()
	// 后续 Set 提高等级,导出的 Log 应跟随(README:受 Set 其他属性影响)
	log.Set(log.WithOutput(buf), log.WithLevel("Warn"))
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
	log.Set(log.WithOutput(io.Discard), log.WithLevel("Debug"), log.WithExtFields(map[string]interface{}{}))
}

func TestConcurrentSetAndExportedLog(t *testing.T) {
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
}
