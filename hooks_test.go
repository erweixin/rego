package rego

import (
	"testing"
)

func TestUse(t *testing.T) {
	runtime := &Runtime{
		refreshChan: make(chan struct{}, 1),
	}
	ctx := newComponentContext("test", nil, runtime)

	// 1. 首次使用状态
	s1 := Use(ctx, "count", 10)
	if s1.Val != 10 {
		t.Errorf("Expected 10, got %v", s1.Val)
	}

	// 2. 更新状态
	s1.Set(20)
	if s1.Val != 20 {
		t.Errorf("Expected 20, got %v", s1.Val)
	}

	// 3. 再次获取状态（模拟重渲染）
	s2 := Use(ctx, "count", 10)
	if s2.Val != 20 {
		t.Errorf("Expected 20, got %v", s2.Val)
	}

	// 4. 函数式更新
	s2.Update(func(old int) int { return old + 5 })
	if s2.Val != 25 {
		t.Errorf("Expected 25, got %v", s2.Val)
	}
}

func TestUseEffect(t *testing.T) {
	runtime := &Runtime{
		refreshChan: make(chan struct{}, 1),
	}
	ctx := newComponentContext("test", nil, runtime)

	runCount := 0
	cleanupCount := 0

	effectFn := func() func() {
		runCount++
		return func() {
			cleanupCount++
		}
	}

	// 1. 首次运行
	UseEffect(ctx, effectFn, "dep1")
	if runCount != 1 {
		t.Errorf("Expected runCount 1, got %d", runCount)
	}

	// 2. 依赖未变，不应运行
	ctx.reset()
	UseEffect(ctx, effectFn, "dep1")
	if runCount != 1 {
		t.Errorf("Expected runCount 1, got %d", runCount)
	}

	// 3. 依赖改变，应执行清理并再次运行
	ctx.reset()
	UseEffect(ctx, effectFn, "dep2")
	if runCount != 2 {
		t.Errorf("Expected runCount 2, got %d", runCount)
	}
	if cleanupCount != 1 {
		t.Errorf("Expected cleanupCount 1, got %d", cleanupCount)
	}

	// 4. 组件卸载时的清理
	ctx.cleanup()
	if cleanupCount != 2 {
		t.Errorf("Expected cleanupCount 2, got %d", cleanupCount)
	}
}

func TestUseMemo(t *testing.T) {
	ctx := newComponentContext("test", nil, nil)

	calcCount := 0
	memoFn := func() int {
		calcCount++
		return calcCount
	}

	// 1. 首次计算
	v1 := UseMemo(ctx, memoFn, "dep1")
	if v1 != 1 {
		t.Errorf("Expected 1, got %d", v1)
	}
	if calcCount != 1 {
		t.Errorf("Expected calcCount 1, got %d", calcCount)
	}

	// 2. 依赖未变，使用缓存
	ctx.reset()
	v2 := UseMemo(ctx, memoFn, "dep1")
	if v2 != 1 {
		t.Errorf("Expected 1, got %d", v2)
	}
	if calcCount != 1 {
		t.Errorf("Expected calcCount 1, got %d", calcCount)
	}

	// 3. 依赖改变，重新计算
	ctx.reset()
	v3 := UseMemo(ctx, memoFn, "dep2")
	if v3 != 2 {
		t.Errorf("Expected 2, got %d", v3)
	}
	if calcCount != 2 {
		t.Errorf("Expected calcCount 2, got %d", calcCount)
	}
}

func TestUseRef(t *testing.T) {
	ctx := newComponentContext("test", nil, nil)

	ref1 := UseRef(ctx, 100)
	if ref1.Current != 100 {
		t.Errorf("Expected 100, got %d", ref1.Current)
	}

	// 模拟修改 ref
	ref1.Current = 200

	// 模拟重渲染，再次获取
	ctx.reset()
	ref2 := UseRef(ctx, 100)
	if ref2.Current != 200 {
		t.Errorf("Expected 200, got %d", ref2.Current)
	}
}
