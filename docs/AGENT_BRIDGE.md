# Rego Bridge: Core-UI Communication Design

This document describes the built-in solution for connecting a UI-agnostic "Agent Core" with the `rego` TUI layer.

## The Problem

Agents usually run in a background loop. They need to:
1. **Stream output**: Send tokens or logs to the UI as they arrive.
2. **Report status**: Update the UI about what they are doing (e.g., "Searching...", "Thinking...").
3. **Request input**: Pause execution to ask the user for confirmation or data.

The UI needs to:
1. **Render state**: Show the current progress.
2. **Provide feedback**: Send user decisions back to the Core.
3. **Control lifecycle**: Start, stop, or pause the Core.

## The Solution: `rego.Bridge`

We introduce `rego.Bridge` which acts as a two-way synchronization point.

### 1. The Core Interface

The Core is designed to receive a "Handle" through which it communicates.

```go
type AgentHandle interface {
    // Send a partial update (token, progress, etc.)
    Update(data any)
    
    // Pause and wait for user interaction
    // Returns the user's response
    Ask(question any) (answer any)
}
```

### 2. The `UseBridge` Hook

In the `rego` UI, you use `UseBridge` to create a controller that manages this communication.

```go
func MyAgentComponent(c rego.C) rego.Node {
    // bridge contains:
    // .State -> The latest data sent via Update()
    // .Interaction -> Current pending request from Ask()
    // .Submit() -> Method to reply to Ask()
    bridge := rego.UseBridge[MyState, MyQuestion, MyAnswer](c, MyState{})

    rego.UseEffect(c, func() func() {
        // Run core logic in background
        go core.Run(bridge.Handle()) 
        return nil
    })

    return rego.VStack(
        // 1. Render streaming content from bridge.State
        rego.Markdown(bridge.State().Text),
        
        // 2. Render interaction UI if core is waiting
        rego.When(bridge.HasInteraction(),
            func() rego.Node {
                q := bridge.Interaction()
                return rego.HStack(
                    rego.Text(q.Prompt),
                    rego.Button(c.Child("confirm"), rego.ButtonProps{
                        Label: "Confirm",
                        OnClick: func() { bridge.Submit(true) },
                    }),
                    rego.Button(c.Child("cancel"), rego.ButtonProps{
                        Label: "Cancel",
                        OnClick: func() { bridge.Submit(false) },
                    }),
                )
            }(),
        ),
    )
}
```

## Detailed Mechanics

### State Synchronization

- `Handle.Update(v)` triggers a `rego.Refresh()` and updates an internal `rego.State`.
- It uses a non-blocking internal channel to ensure the Core doesn't hang if the UI is slow to render.

### Blocking Interaction (The "Wait" Pattern)

1. Core calls `Handle.Ask(q)`.
2. `Bridge` sets its internal `interaction` state and triggers a UI refresh.
3. Core's goroutine blocks on a response channel inside `Ask()`.
4. User interacts with UI, calling `bridge.Submit(a)`.
5. `bridge.Submit` sends the value into the response channel, unblocking the Core.
6. `Bridge` clears the `interaction` state and refreshes UI.

## Benefits

- **Type Safety**: Use Go generics for State and Interactions.
- **Decoupling**: The Agent Core can be tested independently of the UI.
- **DX**: Standardizes how complex Agent workflows (Tool Use, Human-in-the-loop) are implemented in TUIs.

## Complete Example

```go
package main

import (
    "github.com/erweixin/rego"
    "github.com/erweixin/rego/examples/bridge_demo/core"
)

// Define state and interaction types
type AppState struct {
    Status   string
    Progress int
    Logs     []string
}

type Question struct {
    Title string
    Body  string
}

// Create global Context to share Bridge
type AppBridge = *rego.Bridge[AppState, Question, bool]
var BridgeContext = rego.CreateContext[AppBridge](nil)

func App(c rego.C) rego.Node {
    // Create Bridge
    bridge := rego.UseBridge[AppState, Question, bool](c, AppState{Status: "Waiting to start"})

    // Start Core
    rego.UseEffect(c, func() func() {
        go core.Run(bridge.Handle())
        return nil
    })

    // Use Context to share Bridge to all child components
    return BridgeContext.Provide(c, bridge,
        rego.VStack(
            Header(c.Child("header")),
            Workspace(c.Child("workspace")),
            InteractionArea(c.Child("interaction")),
        ),
    )
}

// Deep components can access Bridge via Context
func InteractionArea(c rego.C) rego.Node {
    bridge := rego.UseContext(c, BridgeContext)
    if bridge == nil || !bridge.HasInteraction() {
        return rego.Text("Status: Running...").Dim()
    }

    q := bridge.Interaction()
    return rego.Box(
        rego.VStack(
            rego.Text(q.Title).Bold().Color(rego.Yellow),
            rego.Text(q.Body),
            rego.HStack(
                rego.Button(c.Child("yes"), rego.ButtonProps{
                    Label:   "Confirm",
                    Primary: true,
                    OnClick: func() { bridge.Submit(true) },
                }),
                rego.Button(c.Child("no"), rego.ButtonProps{
                    Label:   "Reject",
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

## API Reference

### UseBridge

```go
func UseBridge[S any, Q any, A any](c C, initial S) *Bridge[S, Q, A]
```

**Type Parameters:**
- `S` - State type
- `Q` - Question type (Agent requests from UI)
- `A` - Answer type (UI replies to Agent)

### Bridge Methods

```go
type Bridge[S, Q, A any] struct { ... }

// Get current state
func (b *Bridge[S, Q, A]) State() S

// Check for pending interaction request
func (b *Bridge[S, Q, A]) HasInteraction() bool

// Get current interaction request
func (b *Bridge[S, Q, A]) Interaction() Q

// Submit answer, unblocking Agent
func (b *Bridge[S, Q, A]) Submit(answer A)

// Get the Agent-side Handle
func (b *Bridge[S, Q, A]) Handle() Handle[S, Q, A]
```

### Handle Interface (Agent-side)

```go
type Handle[S, Q, A any] interface {
    // Update state (non-blocking)
    Update(state S)
    
    // Request user interaction (blocking)
    Ask(question Q) A
}
```
