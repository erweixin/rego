package rego

import (
	"reflect"
)

// =============================================================================
// State 类型
// =============================================================================

// State 表示一个状态值
type State[T any] struct {
	Val T
	ctx *componentContext
	key string
}

// Set 设置状态值并触发重渲染
func (s *State[T]) Set(value T) {
	// 如果值没变，不触发重渲染
	if reflect.DeepEqual(s.Val, value) {
		return
	}
	s.Val = value
	s.ctx.setState(s.key, value)
	s.ctx.Refresh()
}

// Update 使用函数更新状态值
func (s *State[T]) Update(fn func(old T) T) {
	s.Set(fn(s.Val))
}

// =============================================================================
// Use Hook
// =============================================================================

// Use 声明一个状态
func Use[T any](c C, key string, initial T) *State[T] {
	ctx := c.(*componentContext)

	// 检查是否已存在该状态
	if existing, ok := ctx.getState(key); ok {
		return &State[T]{
			Val: existing.(T),
			ctx: ctx,
			key: key,
		}
	}

	// 初始化状态
	ctx.setState(key, initial)
	return &State[T]{
		Val: initial,
		ctx: ctx,
		key: key,
	}
}

// =============================================================================
// UseKey Hook
// =============================================================================

// UseKey 注册键盘事件处理器
func UseKey(c C, handler func(key Key, r rune)) {
	ctx := c.(*componentContext)
	ctx.keyHandler = handler
}

// =============================================================================
// UseMouse Hook
// =============================================================================

// UseMouse 注册鼠标事件处理器
func UseMouse(c C, handler func(ev MouseEvent)) {
	ctx := c.(*componentContext)
	ctx.mouseHandler = handler
}

// =============================================================================
// UseEffect Hook
// =============================================================================

// effectSlot 存储副作用信息
type effectSlot struct {
	deps    []any
	cleanup func()
	ran     bool
}

// UseEffect 声明一个副作用
// fn 返回清理函数，如果不需要清理返回 nil
func UseEffect(c C, fn func() func(), deps ...any) {
	ctx := c.(*componentContext)

	// 生成 effect key
	key := ctx.nextEffectKey()

	// 获取或创建 effect slot
	slot := ctx.getEffectSlot(key)
	if slot == nil {
		slot = &effectSlot{}
		ctx.setEffectSlot(key, slot)
	}

	// 检查依赖是否变化
	shouldRun := !slot.ran || !depsEqual(slot.deps, deps)

	if shouldRun {
		// 执行清理
		if slot.cleanup != nil {
			slot.cleanup()
		}

		// 执行副作用
		cleanup := fn()
		slot.cleanup = cleanup
		slot.deps = deps
		slot.ran = true
	}
}

// depsEqual 比较两个依赖数组是否相等
func depsEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		// 使用 reflect.DeepEqual 而非直接比较，因为依赖项可能包含不可比较类型（如 slice）
		if !reflect.DeepEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

// =============================================================================
// Ref 类型
// =============================================================================

// Ref 引用类型，用于避免闭包陷阱
type Ref[T any] struct {
	Current T
}

// UseRef 创建一个引用
func UseRef[T any](c C, initial T) *Ref[T] {
	ctx := c.(*componentContext)
	key := "__ref__" + string(rune(ctx.refIndex))
	ctx.refIndex++

	if existing, ok := ctx.refs[key]; ok {
		return existing.(*Ref[T])
	}

	ref := &Ref[T]{Current: initial}
	ctx.refs[key] = ref
	return ref
}

// =============================================================================
// UseMemo Hook
// =============================================================================

// memoSlot 存储 memo 信息
type memoSlot struct {
	deps  []any
	value any
}

// UseMemo 缓存计算结果，只在依赖变化时重新计算
func UseMemo[T any](c C, fn func() T, deps ...any) T {
	ctx := c.(*componentContext)

	// 生成 memo key
	key := ctx.nextMemoKey()

	// 获取或创建 memo slot
	slot := ctx.getMemoSlot(key)

	// 如果是首次调用 or 依赖变化，重新计算
	if slot == nil || !depsEqual(slot.deps, deps) {
		value := fn()
		ctx.setMemoSlot(key, &memoSlot{
			deps:  deps,
			value: value,
		})
		return value
	}

	return slot.value.(T)
}
