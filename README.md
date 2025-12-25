# Rego

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/erweixin/rego)](https://goreportcard.com/report/github.com/erweixin/rego)

> Bringing React Hooks-style development experience to Go CLI/TUI

English | [简体中文](README_ZH.md)

---

## Features

- **Hooks Style** - Familiar APIs like `Use`, `UseEffect`, `UseKey`
- **Type Safe** - Built on Go generics with compile-time type checking
- **Explicit Keys** - Break free from React Hooks' call order constraints, use in if/for blocks
- **Declarative UI** - Layout components like `VStack`, `HStack`, `Box`
- **Focus Management** - Built-in Tab/Shift+Tab navigation
- **Mouse Support** - Click, scroll, and hover events
- **Built-in Components** - Button, TextInput, Checkbox, Spinner, Markdown, and more
- **Agent Friendly** - Bridge mechanism, perfect for AI Agent streaming scenarios

---

## Quick Start

### Installation

```bash
go get github.com/erweixin/rego
```

### Hello World

```go
package main

import (
    "fmt"
    "github.com/erweixin/rego"
)

func App(c rego.C) rego.Node {
    count := rego.Use(c, "count", 0)
    
    rego.UseKey(c, func(key rego.Key, r rune) {
        switch r {
        case '+': count.Set(count.Val + 1)
        case '-': count.Set(count.Val - 1)
        case 'q': c.Quit()
        }
    })
    
    return rego.VStack(
        rego.Text("Rego Counter").Bold(),
        rego.Text(fmt.Sprintf("Count: %d", count.Val)),
        rego.Spacer(),
        rego.Text("[+] Increment  [-] Decrement  [q] Quit").Dim(),
    )
}

func main() {
    rego.Run(App)
}
```

Output:

```
Rego Counter
Count: 0

[+] Increment  [-] Decrement  [q] Quit
```

---

## Core Concepts

### Hooks

```go
// State management
count := rego.Use(c, "count", 0)        // Declare state
count.Set(10)                            // Set value
count.Update(func(v int) int { return v + 1 }) // Functional update

// Side effects
rego.UseEffect(c, func() func() {
    ticker := time.NewTicker(time.Second)
    go func() {
        for range ticker.C {
            c.Refresh()
        }
    }()
    return ticker.Stop  // Return cleanup function
}, dep1, dep2)  // Dependency list

// Keyboard events
rego.UseKey(c, func(key rego.Key, r rune) {
    if key == rego.KeyEnter { /* ... */ }
    if r == 'q' { c.Quit() }
})

// Mouse events
rego.UseMouse(c, func(ev rego.MouseEvent) {
    if ev.Type == rego.MouseEventClick { /* ... */ }
})

// Focus management
focus := rego.UseFocus(c)
if focus.IsFocused { /* Current component has focus */ }

// Memoization
result := rego.UseMemo(c, func() int {
    return expensiveCalculation()
}, dep1, dep2)

// Refs (solve closure traps)
ref := rego.UseRef(c, &someValue)
```

### Child Components

```go
func App(c rego.C) rego.Node {
    return rego.VStack(
        Header(c.Child("header")),   // Isolated state space
        Content(c.Child("content")),
        Footer(c.Child("footer")),
    )
}

// Use index in lists
rego.For(items, func(item Item, i int) rego.Node {
    return ItemComponent(c.Child("item", i), item)
})
```

### Layout

```go
// Vertical stack
rego.VStack(
    rego.Text("Title").Bold(),
    rego.Divider(),
    rego.Text("Content"),
    rego.Spacer(),  // Flexible space
    rego.Text("Footer").Dim(),
)

// Horizontal stack
rego.HStack(
    rego.Text("Left"),
    rego.Spacer(),
    rego.Text("Right"),
).Gap(2)

// Container with border
rego.Box(
    rego.Text("Boxed Content"),
).Border(rego.BorderRounded).Padding(1, 2)

// Flex layout
rego.VStack(
    rego.Text("Header").Height(1),
    rego.Box(content).Flex(1),  // Take remaining space
    rego.Text("Footer").Height(1),
)
```

### Styling

```go
rego.Text("Styled Text").
    Bold().
    Italic().
    Underline().
    Color(rego.Cyan).
    Background(rego.Black)

rego.Box(child).
    Border(rego.BorderDouble).
    BorderColor(rego.Green).
    Padding(1, 2).
    Width(40).
    Height(10)
```

---

## Built-in Components

### Button

```go
rego.Button(c.Child("btn"), rego.ButtonProps{
    Label:   "Submit",
    Primary: true,
    OnClick: func() { /* ... */ },
})
```

### TextInput

```go
rego.TextInput(c.Child("input"), rego.TextInputProps{
    Value:       value.Val,
    Placeholder: "Enter text...",
    OnChanged:   func(s string) { value.Set(s) },
    OnSubmit:    func(s string) { /* On Enter */ },
})
```

### Checkbox

```go
rego.Checkbox(c.Child("check"), rego.CheckboxProps{
    Label:     "Accept terms",
    Checked:   agreed.Val,
    OnChanged: func(v bool) { agreed.Set(v) },
})
```

### Spinner

```go
rego.Spinner(c.Child("loading"), "Loading...")
```

### ScrollBox / TailBox

```go
// Scrollable container
rego.ScrollBox(c.Child("scroll"), longContent)

// Auto-scroll to bottom (ideal for logs/chat)
rego.TailBox(c.Child("logs"), logContent)
```

### Markdown

```go
rego.Markdown("# Hello\n\nThis is **markdown** content.")
```

---

## AI Agent Scenarios

Rego is especially suitable for building AI Agent CLIs with built-in Bridge mechanism:

```go
func AgentUI(c rego.C) rego.Node {
    bridge := rego.UseBridge[AgentState, Question, Answer](c, AgentState{})
    
    rego.UseEffect(c, func() func() {
        go agent.Run(bridge.Handle())  // Run Agent in background
        return nil
    })
    
    return rego.VStack(
        // Render streaming output
        rego.Markdown(bridge.State().Response),
        
        // Render interaction requests (e.g., confirmation dialog)
        rego.When(bridge.HasInteraction(),
            ConfirmDialog(c.Child("confirm"), bridge),
        ),
    )
}
```

For more details, see [Agent Bridge Documentation](docs/AGENT_BRIDGE.md).

---

## Examples

| Example | Description |
|---------|-------------|
| [hello](examples/hello) | Simple Hello World |
| [counter](examples/counter) | Counter, demonstrates state management |
| [todo](examples/todo) | Todo app, full feature demo |
| [timer](examples/timer) | Timer, demonstrates UseEffect |
| [focus](examples/focus) | Focus switching, multi-panel app |
| [form](examples/form) | Form, showcases built-in components |
| [dashboard](examples/dashboard) | Dashboard, complex layouts |
| [agent](examples/agent) | AI Agent, streaming output |
| [markdown](examples/markdown) | Markdown rendering |
| [gallery](examples/gallery) | Component gallery, all components |

Run examples:

```bash
cd examples/counter
go run main.go
```

---

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                       User Code                             │
│   func App(c rego.C) rego.Node { ... }                      │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                     rego (Core Layer)                       │
│                                                             │
│  • C Interface - Component Context                          │
│  • State[T] - Generic State Management                      │
│  • Hooks - Use, UseEffect, UseKey, UseMemo, UseRef, UseFocus│
│  • Node - Declarative View Nodes                            │
│  • Context[T] - Cross-component Context Passing             │
│                                                             │
└──────────────────────────┬──────────────────────────────────┘
                           │
┌──────────────────────────▼──────────────────────────────────┐
│                     tcell (Render Layer)                    │
│                                                             │
│  • Terminal Init/Restore                                    │
│  • Screen Rendering + Built-in Diff                         │
│  • Keyboard/Mouse Events                                    │
│  • Cross-platform Support                                   │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Comparison with Other Solutions

| Feature | Rego | bubbletea | tcell |
|---------|------|-----------|-------|
| Architecture | Hooks | Elm (MVU) | Imperative |
| State Management | `rego.Use()` fine-grained | Model centralized | Manual |
| Side Effects | `rego.UseEffect()` | Cmd | Manual |
| State in Conditionals | Yes | - | - |
| Type Safety | Generics | Assertions | - |
| Focus Management | `rego.UseFocus()` | Manual | Manual |
| Learning Curve | React dev friendly | Need to learn Elm | Steep |

---

## Documentation

- [API Reference](docs/API.md) - Complete API documentation
- [Component Guide](docs/COMPONENTS.md) - How to write custom components
- [Styling Guide](docs/STYLING.md) - Colors, borders, layouts explained
- [Agent Bridge](docs/AGENT_BRIDGE.md) - AI Agent integration guide

---

## Development

```bash
# Clone repository
git clone https://github.com/erweixin/rego.git
cd rego

# Run tests
go test ./...

# Run examples
cd examples/gallery
go run .
```

---

## Roadmap

- [x] Core Hooks Runtime
- [x] Basic Layout Components
- [x] Focus Management System
- [x] Mouse Support
- [x] Built-in Component Library
- [x] Markdown Rendering
- [x] Agent Bridge
- [ ] Select/Dropdown Component
- [ ] Table Component
- [ ] Modal Component
- [ ] Theme System

---

## Contributing

Contributions are welcome! Please read the [Contributing Guide](CONTRIBUTING.md).

---

## License

MIT License - See [LICENSE](LICENSE) for details.

---

## Acknowledgments

- [tcell](https://github.com/gdamore/tcell) - Low-level terminal library
- [glamour](https://github.com/charmbracelet/glamour) - Markdown rendering
- [React Hooks](https://react.dev/reference/react) - Design inspiration

---

<p align="center">
  <sub>Made with love for the Go community</sub>
</p>
