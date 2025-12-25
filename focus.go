package rego

import "sync"

// =============================================================================
// FocusState - 组件的焦点状态
// =============================================================================

// FocusState 表示组件的焦点状态
type FocusState struct {
	IsFocused bool   // 当前是否有焦点
	Focus     func() // 获取焦点
	Blur      func() // 失去焦点
	ctx       *componentContext
}

// =============================================================================
// FocusManager - 全局焦点管理器
// =============================================================================

// FocusManager 管理所有可聚焦组件
type FocusManager struct {
	mu         sync.RWMutex
	focusable  []string                     // 可聚焦组件的 key 列表（有序）
	focusMap   map[string]*componentContext // key -> context
	currentKey string                       // 当前聚焦的组件 key
	order      int                          // 注册顺序计数器
	orderMap   map[string]int               // key -> 注册顺序
}

// newFocusManager 创建焦点管理器
func newFocusManager() *FocusManager {
	return &FocusManager{
		focusMap: make(map[string]*componentContext),
		orderMap: make(map[string]int),
	}
}

// Register 注册可聚焦组件
func (fm *FocusManager) Register(key string, ctx *componentContext) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// 检查是否已注册
	if _, exists := fm.focusMap[key]; exists {
		return
	}

	fm.focusMap[key] = ctx
	fm.orderMap[key] = fm.order
	fm.order++

	// 按注册顺序插入
	fm.focusable = append(fm.focusable, key)

	// 如果还没有焦点，自动聚焦到第一个组件
	if fm.currentKey == "" {
		fm.currentKey = key
	}
}

// Unregister 注销可聚焦组件
func (fm *FocusManager) Unregister(key string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	delete(fm.focusMap, key)
	delete(fm.orderMap, key)

	// 从列表中移除
	for i, k := range fm.focusable {
		if k == key {
			fm.focusable = append(fm.focusable[:i], fm.focusable[i+1:]...)
			break
		}
	}

	// 如果当前聚焦的组件被移除，切换到下一个
	if fm.currentKey == key {
		if len(fm.focusable) > 0 {
			fm.currentKey = fm.focusable[0]
		} else {
			fm.currentKey = ""
		}
	}
}

// Focus 聚焦到指定组件
func (fm *FocusManager) Focus(key string) {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if _, exists := fm.focusMap[key]; exists {
		fm.currentKey = key
	}
}

// Current 获取当前聚焦的组件 key
func (fm *FocusManager) Current() string {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.currentKey
}

// CurrentContext 获取当前聚焦的组件上下文
func (fm *FocusManager) CurrentContext() *componentContext {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.focusMap[fm.currentKey]
}

// IsFocused 检查指定组件是否有焦点
func (fm *FocusManager) IsFocused(key string) bool {
	fm.mu.RLock()
	defer fm.mu.RUnlock()
	return fm.currentKey == key
}

// Next 切换到下一个可聚焦组件
func (fm *FocusManager) Next() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if len(fm.focusable) == 0 {
		return
	}

	// 找到当前索引
	currentIdx := -1
	for i, key := range fm.focusable {
		if key == fm.currentKey {
			currentIdx = i
			break
		}
	}

	// 切换到下一个
	nextIdx := (currentIdx + 1) % len(fm.focusable)
	fm.currentKey = fm.focusable[nextIdx]
}

// Prev 切换到上一个可聚焦组件
func (fm *FocusManager) Prev() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	if len(fm.focusable) == 0 {
		return
	}

	// 找到当前索引
	currentIdx := -1
	for i, key := range fm.focusable {
		if key == fm.currentKey {
			currentIdx = i
			break
		}
	}

	// 切换到上一个
	prevIdx := currentIdx - 1
	if prevIdx < 0 {
		prevIdx = len(fm.focusable) - 1
	}
	fm.currentKey = fm.focusable[prevIdx]
}

// Reset 重置焦点管理器（每次渲染前调用）
func (fm *FocusManager) Reset() {
	fm.mu.Lock()
	defer fm.mu.Unlock()

	// 清空注册列表，但保留当前焦点
	fm.focusable = fm.focusable[:0]
	fm.focusMap = make(map[string]*componentContext)
	fm.orderMap = make(map[string]int)
	fm.order = 0
}

// =============================================================================
// UseFocus Hook
// =============================================================================

// UseFocus 声明组件可聚焦，返回焦点状态
func UseFocus(c C) FocusState {
	ctx := c.(*componentContext)
	runtime := ctx.runtime

	if runtime == nil || runtime.focusManager == nil {
		return FocusState{IsFocused: false}
	}

	fm := runtime.focusManager

	// 生成唯一的焦点 key（基于组件路径）
	focusKey := ctx.focusKey()

	// 注册为可聚焦组件
	fm.Register(focusKey, ctx)

	// 自动集成鼠标点击聚焦
	ctx.mouseHandler = func(ev MouseEvent) {
		if ev.Type == MouseEventClick && ev.Button == MouseButtonLeft {
			// 只有点击在组件范围内时才聚焦
			if ctx.Rect().Contains(ev.X, ev.Y) {
				fm.Focus(focusKey)
				ctx.Refresh()
			}
		}
	}

	// 返回焦点状态
	return FocusState{
		IsFocused: fm.IsFocused(focusKey),
		Focus: func() {
			fm.Focus(focusKey)
			ctx.Refresh()
		},
		Blur: func() {
			// Blur 时切换到下一个
			if fm.IsFocused(focusKey) {
				fm.Next()
				ctx.Refresh()
			}
		},
		ctx: ctx,
	}
}

// focusKey 生成组件的焦点 key（基于组件路径）
func (c *componentContext) focusKey() string {
	if c.parent == nil {
		return c.key
	}
	return c.parent.focusKey() + "/" + c.key
}
