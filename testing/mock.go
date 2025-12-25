package testing

import (
	"github.com/erweixin/rego"
	"github.com/gdamore/tcell/v2"
)

// MockScreen 包装了 tcell.SimulationScreen，提供更方便的测试接口
type MockScreen struct {
	tcell.SimulationScreen
}

// NewMockScreen 创建一个指定大小的模拟屏幕
func NewMockScreen(w, h int) *MockScreen {
	s := tcell.NewSimulationScreen("")
	s.Init()
	s.SetSize(w, h)
	return &MockScreen{s}
}

// GetContentString 获取屏幕内容的字符串表示（忽略样式）
func (s *MockScreen) GetContentString() string {
	w, h := s.Size()
	var res string
	for y := 0; y < h; y++ {
		line := ""
		for x := 0; x < w; x++ {
			r, _, _, _ := s.GetContent(x, y)
			if r == 0 {
				line += " "
			} else {
				line += string(r)
			}
		}
		res += line
		if y < h-1 {
			res += "\n"
		}
	}
	return res
}

// TestRuntime 提供一个用于测试的 Runtime 包装
type TestRuntime struct {
	*rego.Runtime
	Screen *MockScreen
}

// NewTestRuntime 创建一个用于测试的运行时
func NewTestRuntime(root func(rego.C) rego.Node, w, h int) *TestRuntime {
	screen := NewMockScreen(w, h)
	r := rego.NewTestRuntime(root, screen)
	return &TestRuntime{
		Runtime: r,
		Screen:  screen,
	}
}
