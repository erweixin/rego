package rego

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestPanicRecovery(t *testing.T) {
	// 创建一个模拟屏幕
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to init simulation screen: %v", err)
	}

	// 创建一个会 panic 的运行时
	runtime := newRuntime(func(c C) Node {
		panic("oops")
	})
	runtime.screen = screen
	runtime.rootContext = newComponentContext("root", nil, runtime)

	// 执行渲染，应该捕获 panic
	runtime.render()

	if runtime.lastPanic == nil {
		t.Error("expected lastPanic to be set, but it was nil")
	}

	if runtime.lastPanic != "oops" {
		t.Errorf("expected lastPanic 'oops', got %v", runtime.lastPanic)
	}

	// 验证屏幕上是否绘制了错误信息
	// "REGO RUNTIME PANIC" 应该在屏幕上
	found := false
	w, h := screen.Size()
	for y := 0; y < h; y++ {
		line := ""
		for x := 0; x < w; x++ {
			mainc, _, _, _ := screen.GetContent(x, y)
			line += string(mainc)
		}
		if contains(line, "REGO RUNTIME PANIC") {
			found = true
			break
		}
	}

	if !found {
		t.Error("could not find panic message on error screen")
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
