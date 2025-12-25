package rego

import (
	"fmt"
	"runtime/debug"

	"github.com/gdamore/tcell/v2"
)

// Runtime 是应用运行时
type Runtime struct {
	screen       tcell.Screen
	root         func(C) Node
	rootContext  *componentContext
	focusManager *FocusManager

	refreshChan chan struct{}
	quitChan    chan struct{}

	// 光标位置（用于 IME 输入定位）
	cursorX, cursorY int
	showCursor       bool

	// 错误处理
	lastPanic  any
	panicStack []byte
}

// newRuntime 创建运行时
func newRuntime(root func(C) Node) *Runtime {
	return &Runtime{
		root:         root,
		focusManager: newFocusManager(),
		refreshChan:  make(chan struct{}, 1),
		quitChan:     make(chan struct{}),
	}
}

// Run 启动运行时
func (r *Runtime) Run() error {
	// 初始化 tcell screen
	screen, err := tcell.NewScreen()
	if err != nil {
		return err
	}
	if err := screen.Init(); err != nil {
		return err
	}
	defer func() {
		// 清理所有 effects
		if r.rootContext != nil {
			r.rootContext.cleanup()
		}
		screen.Fini()
	}()

	r.screen = screen
	r.rootContext = newComponentContext("root", nil, r)

	// 启用粘贴模式（改善 IME 支持）
	screen.EnablePaste()

	// 隐藏光标（避免 IME 定位问题）
	screen.HideCursor()

	// 启用鼠标支持（包含运动追踪以支持 Hover）
	screen.EnableMouse(tcell.MouseButtonEvents | tcell.MouseMotionEvents)

	// 初始渲染
	r.render()

	// 启动事件监听协程
	eventChan := make(chan tcell.Event)
	go func() {
		for {
			ev := screen.PollEvent()
			if ev == nil {
				return
			}
			eventChan <- ev
		}
	}()

	// 主循环
	for {
		select {
		case <-r.quitChan:
			return nil

		case <-r.refreshChan:
			r.render()

		case ev := <-eventChan:
			r.handleEvent(ev)
		}
	}
}

// render 执行渲染
func (r *Runtime) render() {
	r.screen.Clear()

	// 如果之前发生了 panic，显示错误界面
	if r.lastPanic != nil {
		r.drawErrorScreen()
		r.screen.Show()
		return
	}

	defer func() {
		if err := recover(); err != nil {
			r.lastPanic = err
			r.panicStack = debug.Stack()
			r.screen.Clear()
			r.drawErrorScreen()
			r.screen.Show()
		}
	}()

	r.rootContext.reset()

	// 重置焦点管理器（每次渲染前）
	r.focusManager.Reset()

	// 重置光标状态（每次渲染前）
	r.showCursor = false

	// 调用根组件
	node := r.root(r.rootContext)

	// 准备渲染屏幕代理（拦截光标设置）
	renderScreen := &renderScreenProxy{
		Screen:  r.screen,
		runtime: r,
	}

	// 渲染到屏幕
	width, height := r.screen.Size()
	if node != nil {
		node.render(renderScreen, 0, 0, width, height)
	}

	// 设置光标位置（用于 IME 输入定位）
	if r.showCursor {
		r.screen.ShowCursor(r.cursorX, r.cursorY)
	} else {
		r.screen.HideCursor()
	}

	r.screen.Show()
}

// renderScreenProxy 代理 tcell.Screen 以拦截光标设置
type renderScreenProxy struct {
	tcell.Screen
	runtime *Runtime
}

func (p *renderScreenProxy) ShowCursor(x, y int) {
	if p.runtime != nil {
		p.runtime.setCursor(x, y)
	}
}

func (p *renderScreenProxy) HideCursor() {
	if p.runtime != nil {
		p.runtime.showCursor = false
	}
}

// handleEvent 处理事件
func (r *Runtime) handleEvent(event tcell.Event) {
	switch e := event.(type) {
	case *tcell.EventKey:
		// Ctrl+C 退出
		if e.Key() == tcell.KeyCtrlC {
			r.quit()
			return
		}

		// Tab/Shift+Tab 焦点导航
		if e.Key() == tcell.KeyTab {
			if e.Modifiers()&tcell.ModShift != 0 {
				r.focusManager.Prev()
			} else {
				r.focusManager.Next()
			}
			r.scheduleRefresh()
			return
		}

		// 转换按键
		key, ru, _ := convertTcellKey(e)

		// 分发给组件树
		r.rootContext.dispatchKeyEvent(key, ru)

	case *tcell.EventMouse:
		ev := convertTcellMouseEvent(e)
		r.rootContext.dispatchMouseEvent(ev)

	case *tcell.EventResize:
		r.scheduleRefresh()
	}
}

// convertTcellMouseEvent 将 tcell 鼠标事件转换为 rego 鼠标事件
func convertTcellMouseEvent(e *tcell.EventMouse) MouseEvent {
	x, y := e.Position()
	button := MouseButtonNone
	eventType := MouseEventMove // 默认为移动

	b := e.Buttons()
	if b&tcell.Button1 != 0 {
		button = MouseButtonLeft
		eventType = MouseEventClick
	} else if b&tcell.Button3 != 0 {
		button = MouseButtonRight
		eventType = MouseEventClick
	} else if b&tcell.Button2 != 0 {
		button = MouseButtonMiddle
		eventType = MouseEventClick
	}

	// 处理滚轮
	if b&tcell.WheelUp != 0 {
		eventType = MouseEventScrollUp
	} else if b&tcell.WheelDown != 0 {
		eventType = MouseEventScrollDown
	}

	return MouseEvent{
		X:      x,
		Y:      y,
		Button: button,
		Type:   eventType,
	}
}

// scheduleRefresh 调度刷新
func (r *Runtime) scheduleRefresh() {
	select {
	case r.refreshChan <- struct{}{}:
	default:
		// 已有刷新请求，忽略
	}
}

// quit 退出应用
func (r *Runtime) quit() {
	close(r.quitChan)
}

// setCursor 设置光标位置
func (r *Runtime) setCursor(x, y int) {
	r.cursorX = x
	r.cursorY = y
	r.showCursor = true
}

// drawErrorScreen 绘制错误界面
func (r *Runtime) drawErrorScreen() {
	w, h := r.screen.Size()
	style := tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite)

	// 填充红色背景
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r.screen.SetContent(x, y, ' ', nil, style)
		}
	}

	// 绘制标题
	title := "  REGO RUNTIME PANIC  "
	r.drawText((w-len(title))/2, 2, title, style.Bold(true))

	// 绘制错误信息
	msg := fmt.Sprintf("Error: %v", r.lastPanic)
	r.drawText(2, 4, msg, style)

	// 绘制堆栈（截取部分）
	r.drawText(2, 6, "Stack Trace:", style.Underline(true))
	lines := fmt.Sprintf("%s", r.panicStack)
	for i, line := range splitLines(lines) {
		if i > h-10 {
			break
		}
		r.drawText(2, 7+i, line, style)
	}

	// 绘制退出提示
	footer := "Press Ctrl+C to quit"
	r.drawText((w-len(footer))/2, h-2, footer, style.Dim(true))
}

func (r *Runtime) drawText(x, y int, text string, style tcell.Style) {
	for i, ru := range text {
		r.screen.SetContent(x+i, y, ru, nil, style)
	}
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
