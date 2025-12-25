package rego

import (
	"sync"
)

// Bridge 是 UI 侧持有的句柄，用于与 Core 通信
type Bridge[S any, Q any, A any] struct {
	ctx         C
	state       *State[S]
	interaction *State[*pendingInteraction[Q, A]]
	handle      *bridgeHandle[S, Q, A]
}

type pendingInteraction[Q any, A any] struct {
	question Q
	answerCh chan A
}

// State 返回当前从 Core 同步过来的状态
func (b *Bridge[S, Q, A]) State() S {
	return b.state.Val
}

// HasInteraction 检查是否有挂起的交互请求
func (b *Bridge[S, Q, A]) HasInteraction() bool {
	return b.interaction.Val != nil
}

// Interaction 返回当前的交互请求内容
func (b *Bridge[S, Q, A]) Interaction() Q {
	if b.interaction.Val == nil {
		var zero Q
		return zero
	}
	return b.interaction.Val.question
}

// Submit 提交用户的回答，解除 Core 的阻塞
func (b *Bridge[S, Q, A]) Submit(answer A) {
	if b.interaction.Val != nil {
		b.interaction.Val.answerCh <- answer
		b.interaction.Set(nil)
	}
}

// Handle 返回给 Core 使用的句柄
func (b *Bridge[S, Q, A]) Handle() Handle[S, Q, A] {
	return b.handle
}

// Handle 是 Core 侧持有的接口
type Handle[S any, Q any, A any] interface {
	Update(state S)
	Ask(question Q) A
}

type bridgeHandle[S any, Q any, A any] struct {
	bridge *Bridge[S, Q, A]
	mu     sync.Mutex
}

func (h *bridgeHandle[S, Q, A]) Update(state S) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.bridge.state.Set(state)
}

func (h *bridgeHandle[S, Q, A]) Ask(question Q) A {
	answerCh := make(chan A)

	// 在 UI 线程设置交互状态
	h.bridge.interaction.Set(&pendingInteraction[Q, A]{
		question: question,
		answerCh: answerCh,
	})

	// 阻塞等待 UI 侧通过 Submit 传回结果
	return <-answerCh
}

// UseBridge 创建一个双向通信桥梁
// S: 状态类型, Q: 问题类型, A: 回答类型
func UseBridge[S any, Q any, A any](c C, initial S) *Bridge[S, Q, A] {
	// 显式获取状态，确保在当前上下文存在
	state := Use(c, "bridge_state", initial)
	interaction := Use[*pendingInteraction[Q, A]](c, "bridge_interaction", nil)

	// 我们每次都创建一个新的 Bridge 包装对象，但它内部引用的 state 是持久的
	// 这样可以避免 UseMemo 闭包捕获带来的潜在引用问题
	b := &Bridge[S, Q, A]{
		ctx:         c,
		state:       state,
		interaction: interaction,
	}
	b.handle = &bridgeHandle[S, Q, A]{bridge: b}
	return b
}
