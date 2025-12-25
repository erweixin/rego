package rego

import (
	"fmt"
	"sync"
)

// Rect 表示屏幕上的矩形区域
type Rect struct {
	X, Y, W, H int
}

// Contains 检查点是否在矩形内
func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H
}

// MouseEventType 鼠标事件类型
type MouseEventType int

const (
	MouseEventPress MouseEventType = iota
	MouseEventRelease
	MouseEventClick
	MouseEventMove
	MouseEventScrollUp
	MouseEventScrollDown
)

// MouseButton 鼠标按钮
type MouseButton int

const (
	MouseButtonNone MouseButton = iota
	MouseButtonLeft
	MouseButtonMiddle
	MouseButtonRight
)

// MouseEvent 鼠标事件
type MouseEvent struct {
	X, Y   int
	Button MouseButton
	Type   MouseEventType
}

// C 是组件上下文接口
type C interface {
	// Child 获取子组件上下文
	Child(key string, index ...int) C

	// Refresh 手动触发重渲染
	Refresh()

	// Quit 退出应用
	Quit()

	// SetCursor 设置光标位置（用于 IME 输入定位）
	SetCursor(x, y int)

	// Wrap 包装节点以追踪其位置（用于鼠标点击）
	Wrap(node Node) *componentNode

	// Rect 获取当前组件的屏幕区域
	Rect() Rect
}

// =============================================================================
// componentContext 实现
// =============================================================================

type componentContext struct {
	key      string
	parent   *componentContext
	children map[string]*componentContext

	// 布局追踪
	rect Rect

	// 状态存储
	states map[string]any

	// Effect 存储
	effects     map[int]*effectSlot
	effectIndex int

	// Ref 存储
	refs     map[string]any
	refIndex int

	// Memo 存储
	memos     map[int]*memoSlot
	memoIndex int

	// Context 值存储
	contextValues map[string]any

	// 事件处理器
	keyHandler   func(Key, rune)
	mouseHandler func(MouseEvent)

	// 运行时引用
	runtime *Runtime

	mu sync.RWMutex
}

func newComponentContext(key string, parent *componentContext, runtime *Runtime) *componentContext {
	return &componentContext{
		key:      key,
		parent:   parent,
		children: make(map[string]*componentContext),
		states:   make(map[string]any),
		effects:  make(map[int]*effectSlot),
		refs:     make(map[string]any),
		memos:    make(map[int]*memoSlot),
		runtime:  runtime,
	}
}

func (c *componentContext) Child(key string, index ...int) C {
	fullKey := key
	if len(index) > 0 {
		fullKey = fmt.Sprintf("%s[%d]", key, index[0])
	}

	if child, ok := c.children[fullKey]; ok {
		child.reset()
		return child
	}

	child := newComponentContext(fullKey, c, c.runtime)
	c.children[fullKey] = child
	return child
}

func (c *componentContext) Refresh() {
	if c.runtime != nil {
		c.runtime.scheduleRefresh()
	}
}

func (c *componentContext) Quit() {
	if c.runtime != nil {
		c.runtime.quit()
	}
}

func (c *componentContext) SetCursor(x, y int) {
	if c.runtime != nil {
		c.runtime.setCursor(x, y)
	}
}

func (c *componentContext) Wrap(node Node) *componentNode {
	return &componentNode{ctx: c, node: node}
}

func (c *componentContext) Rect() Rect {
	return c.rect
}

// reset 重置组件状态索引（每次渲染前调用）
func (c *componentContext) reset() {
	c.effectIndex = 0
	c.refIndex = 0
	c.memoIndex = 0
	c.keyHandler = nil
	c.mouseHandler = nil
}

// getState 获取状态值
func (c *componentContext) getState(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.states[key]
	return v, ok
}

// setState 设置状态值
func (c *componentContext) setState(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.states[key] = value
}

// nextEffectKey 获取下一个 effect key
func (c *componentContext) nextEffectKey() int {
	key := c.effectIndex
	c.effectIndex++
	return key
}

// getEffectSlot 获取 effect slot
func (c *componentContext) getEffectSlot(key int) *effectSlot {
	return c.effects[key]
}

// setEffectSlot 设置 effect slot
func (c *componentContext) setEffectSlot(key int, slot *effectSlot) {
	c.effects[key] = slot
}

// nextMemoKey 获取下一个 memo key
func (c *componentContext) nextMemoKey() int {
	key := c.memoIndex
	c.memoIndex++
	return key
}

// getMemoSlot 获取 memo slot
func (c *componentContext) getMemoSlot(key int) *memoSlot {
	return c.memos[key]
}

// setMemoSlot 设置 memo slot
func (c *componentContext) setMemoSlot(key int, slot *memoSlot) {
	c.memos[key] = slot
}

// dispatchKeyEvent 分发键盘事件（广播模式：所有 handler 都会收到）
func (c *componentContext) dispatchKeyEvent(key Key, r rune) {
	// 1. 自己先处理（父组件优先，处理全局快捷键如 Tab）
	if c.keyHandler != nil {
		c.keyHandler(key, r)
	}

	// 2. 再分发给子组件
	for _, child := range c.children {
		child.dispatchKeyEvent(key, r)
	}
}

// dispatchMouseEvent 分发鼠标事件
func (c *componentContext) dispatchMouseEvent(ev MouseEvent) {
	// 检查事件是否落在自己的矩形区域内
	inRect := c.rect.Contains(ev.X, ev.Y)

	if c.mouseHandler != nil {
		// 即使不在区域内也发送事件，让 handler 自己决定是否处理 (比如用于 MouseLeave)
		// 但为了方便，我们可以在 ev 中标记是否在区域内
		evIn := ev
		if !inRect {
			// 如果不在区域内，且类型是 Move，这通常意味着 MouseLeave
			// 或者简单的“非我区域的移动”
			evIn.Type = MouseEventMove
		}
		c.mouseHandler(evIn)
	}

	// 2. 递归分发给子组件
	for _, child := range c.children {
		child.dispatchMouseEvent(ev)
	}
}

// cleanup 清理所有 effects
func (c *componentContext) cleanup() {
	for _, slot := range c.effects {
		if slot.cleanup != nil {
			slot.cleanup()
		}
	}
	for _, child := range c.children {
		child.cleanup()
	}
}
