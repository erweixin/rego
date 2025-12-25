// Package rego 提供 React Hooks 风格的 CLI/TUI 开发体验
package rego

import "github.com/gdamore/tcell/v2"

// Run 启动应用
func Run(root func(C) Node) error {
	runtime := newRuntime(root)
	return runtime.Run()
}

// NewTestRuntime 创建一个用于测试的运行时
func NewTestRuntime(root func(C) Node, screen tcell.Screen) *Runtime {
	r := newRuntime(root)
	r.screen = screen
	r.rootContext = newComponentContext("root", nil, r)
	return r
}

// Render 立即执行一次渲染（用于测试）
func (r *Runtime) Render() {
	r.render()
}

// DispatchKey 分发键盘事件（用于测试）
func (r *Runtime) DispatchKey(key tcell.Key, r_rune rune, mod tcell.ModMask) {
	ev := tcell.NewEventKey(key, r_rune, mod)
	r.handleEvent(ev)
}
