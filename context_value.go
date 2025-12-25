package rego

import "sync"

// =============================================================================
// Context - 跨组件状态共享
// =============================================================================

// Context 表示一个可跨组件共享的上下文
type Context[T any] struct {
	key          string
	defaultValue T
}

// contextKey 用于在 componentContext 中存储 context 值
type contextKey string

// CreateContext 创建一个新的 Context
func CreateContext[T any](defaultValue T) *Context[T] {
	return &Context[T]{
		key:          generateContextKey(),
		defaultValue: defaultValue,
	}
}

// contextKeyCounter 用于生成唯一的 context key
var (
	contextKeyCounter int
	contextKeyMu      sync.Mutex
)

func generateContextKey() string {
	contextKeyMu.Lock()
	defer contextKeyMu.Unlock()
	contextKeyCounter++
	return string(rune('A' + contextKeyCounter - 1))
}

// Provide 提供 Context 值，包装子节点
// 用法: ThemeContext.Provide(c, "dark", child1, child2, ...)
func (ctx *Context[T]) Provide(c C, value T, children ...Node) Node {
	cc := c.(*componentContext)

	// 存储值到当前组件上下文
	cc.setContextValue(ctx.key, value)

	// 返回包含所有子节点的 VStack
	if len(children) == 0 {
		return Empty()
	}
	if len(children) == 1 {
		return children[0]
	}
	return VStack(children...)
}

// ProvideH 提供 Context 值，子节点水平排列
func (ctx *Context[T]) ProvideH(c C, value T, children ...Node) Node {
	cc := c.(*componentContext)
	cc.setContextValue(ctx.key, value)

	if len(children) == 0 {
		return Empty()
	}
	if len(children) == 1 {
		return children[0]
	}
	return HStack(children...)
}

// =============================================================================
// UseContext Hook
// =============================================================================

// UseContext 获取 Context 的值
// 从当前组件向上查找，直到找到 Provider 或返回默认值
func UseContext[T any](c C, ctx *Context[T]) T {
	cc := c.(*componentContext)

	// 从当前组件向上查找
	for current := cc; current != nil; current = current.parent {
		if value, ok := current.getContextValue(ctx.key); ok {
			return value.(T)
		}
	}

	// 未找到，返回默认值
	return ctx.defaultValue
}

// =============================================================================
// componentContext 扩展
// =============================================================================

// 在 componentContext 中添加 context 值存储
// 注意：这些方法需要添加到 context.go 中

// setContextValue 设置 context 值
func (c *componentContext) setContextValue(key string, value any) {
	if c.contextValues == nil {
		c.contextValues = make(map[string]any)
	}
	c.contextValues[key] = value
}

// getContextValue 获取 context 值
func (c *componentContext) getContextValue(key string) (any, bool) {
	if c.contextValues == nil {
		return nil, false
	}
	v, ok := c.contextValues[key]
	return v, ok
}
