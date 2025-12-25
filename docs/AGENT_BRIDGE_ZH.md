# Rego Bridge: Core-UI 通信设计

本文档描述了连接 UI 无关的 "Agent Core" 与 `rego` TUI 层的内置解决方案。

## 问题背景

Agent 通常在后台循环中运行。它们需要：
1. **流式输出**: 在 Token 或日志到达时发送到 UI。
2. **报告状态**: 更新 UI 关于当前正在做什么（如 "搜索中..."、"思考中..."）。
3. **请求输入**: 暂停执行以询问用户确认或数据。

UI 需要：
1. **渲染状态**: 显示当前进度。
2. **提供反馈**: 将用户决策发送回 Core。
3. **控制生命周期**: 启动、停止或暂停 Core。

## 解决方案: `rego.Bridge`

我们引入 `rego.Bridge` 作为双向同步点。

### 1. Core 接口

Core 设计为接收一个 "Handle"，通过它进行通信。

```go
type AgentHandle interface {
    // 发送部分更新（Token、进度等）
    Update(data any)
    
    // 暂停并等待用户交互
    // 返回用户的响应
    Ask(question any) (answer any)
}
```

### 2. `UseBridge` Hook

在 `rego` UI 中，使用 `UseBridge` 创建一个管理此通信的控制器。

```go
func MyAgentComponent(c rego.C) rego.Node {
    // bridge 包含:
    // .State -> 通过 Update() 发送的最新数据
    // .Interaction -> 来自 Ask() 的当前待处理请求
    // .Submit() -> 回复 Ask() 的方法
    bridge := rego.UseBridge[MyState, MyQuestion, MyAnswer](c, MyState{})

    rego.UseEffect(c, func() func() {
        // 在后台运行 core 逻辑
        go core.Run(bridge.Handle()) 
        return nil
    })

    return rego.VStack(
        // 1. 从 bridge.State 渲染流式内容
        rego.Markdown(bridge.State().Text),
        
        // 2. 如果 core 正在等待，渲染交互 UI
        rego.When(bridge.HasInteraction(),
            func() rego.Node {
                q := bridge.Interaction()
                return rego.HStack(
                    rego.Text(q.Prompt),
                    rego.Button(c.Child("confirm"), rego.ButtonProps{
                        Label: "确认",
                        OnClick: func() { bridge.Submit(true) },
                    }),
                    rego.Button(c.Child("cancel"), rego.ButtonProps{
                        Label: "取消",
                        OnClick: func() { bridge.Submit(false) },
                    }),
                )
            }(),
        ),
    )
}
```

## 详细机制

### 状态同步

- `Handle.Update(v)` 触发 `rego.Refresh()` 并更新内部 `rego.State`。
- 它使用非阻塞的内部 channel，确保 Core 不会在 UI 渲染较慢时挂起。

### 阻塞交互（"等待"模式）

1. Core 调用 `Handle.Ask(q)`。
2. `Bridge` 设置其内部 `interaction` 状态并触发 UI 刷新。
3. Core 的 goroutine 在 `Ask()` 内部的响应 channel 上阻塞。
4. 用户与 UI 交互，调用 `bridge.Submit(a)`。
5. `bridge.Submit` 将值发送到响应 channel，解除 Core 的阻塞。
6. `Bridge` 清除 `interaction` 状态并刷新 UI。

## 优势

- **类型安全**: 使用 Go 泛型处理 State 和 Interactions。
- **解耦**: Agent Core 可以独立于 UI 进行测试。
- **开发体验**: 标准化了复杂 Agent 工作流（工具使用、人机协作）在 TUI 中的实现方式。

## 完整示例

```go
package main

import (
    "github.com/erweixin/rego"
    "github.com/erweixin/rego/examples/bridge_demo/core"
)

// 定义状态和交互类型
type AppState struct {
    Status   string
    Progress int
    Logs     []string
}

type Question struct {
    Title string
    Body  string
}

// 创建全局 Context 共享 Bridge
type AppBridge = *rego.Bridge[AppState, Question, bool]
var BridgeContext = rego.CreateContext[AppBridge](nil)

func App(c rego.C) rego.Node {
    // 创建 Bridge
    bridge := rego.UseBridge[AppState, Question, bool](c, AppState{Status: "等待启动"})

    // 启动 Core
    rego.UseEffect(c, func() func() {
        go core.Run(bridge.Handle())
        return nil
    })

    // 使用 Context 共享 Bridge 到所有子组件
    return BridgeContext.Provide(c, bridge,
        rego.VStack(
            Header(c.Child("header")),
            Workspace(c.Child("workspace")),
            InteractionArea(c.Child("interaction")),
        ),
    )
}

// 深层组件可以通过 Context 获取 Bridge
func InteractionArea(c rego.C) rego.Node {
    bridge := rego.UseContext(c, BridgeContext)
    if bridge == nil || !bridge.HasInteraction() {
        return rego.Text("状态：运行中...").Dim()
    }

    q := bridge.Interaction()
    return rego.Box(
        rego.VStack(
            rego.Text(q.Title).Bold().Color(rego.Yellow),
            rego.Text(q.Body),
            rego.HStack(
                rego.Button(c.Child("yes"), rego.ButtonProps{
                    Label:   "确认",
                    Primary: true,
                    OnClick: func() { bridge.Submit(true) },
                }),
                rego.Button(c.Child("no"), rego.ButtonProps{
                    Label:   "拒绝",
                    OnClick: func() { bridge.Submit(false) },
                }),
            ),
        ),
    ).Border(rego.BorderRounded).BorderColor(rego.Yellow).Padding(1, 2)
}

func main() {
    rego.Run(App)
}
```

## API 参考

### UseBridge

```go
func UseBridge[S any, Q any, A any](c C, initial S) *Bridge[S, Q, A]
```

**类型参数:**
- `S` - 状态类型
- `Q` - 问题类型（Agent 向 UI 请求）
- `A` - 答案类型（UI 回复给 Agent）

### Bridge 方法

```go
type Bridge[S, Q, A any] struct { ... }

// 获取当前状态
func (b *Bridge[S, Q, A]) State() S

// 检查是否有待处理的交互请求
func (b *Bridge[S, Q, A]) HasInteraction() bool

// 获取当前交互请求
func (b *Bridge[S, Q, A]) Interaction() Q

// 提交答案，解除 Agent 阻塞
func (b *Bridge[S, Q, A]) Submit(answer A)

// 获取 Agent 端的 Handle
func (b *Bridge[S, Q, A]) Handle() Handle[S, Q, A]
```

### Handle 接口（Agent 端）

```go
type Handle[S, Q, A any] interface {
    // 更新状态（非阻塞）
    Update(state S)
    
    // 请求用户交互（阻塞）
    Ask(question Q) A
}
```

