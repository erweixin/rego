# Rego Component Development Guide

> How to create and compose custom components

---

## Core Concepts

In Rego, **a component is just a function**:

```go
func ComponentName(c rego.C, args...) rego.Node {
    // Use Hooks
    // Return Node
}
```

- First parameter must be `c rego.C` (component context)
- Return type is `rego.Node` (view node)
- Can have any number of additional parameters
- Hooks use package-level functions: `rego.Use(c, ...)` not `c.Use(...)`

---

## Basic Components

### Stateless Components

```go
func Divider(c rego.C) rego.Node {
    return rego.Text("────────────────────").Dim()
}

func Title(c rego.C, text string) rego.Node {
    return rego.Text(text).Bold().Color(rego.Cyan)
}
```

### Stateful Components

```go
func Counter(c rego.C) rego.Node {
    // Package-level generic function, type-safe
    count := rego.Use(c, "count", 0)
    
    // Handle events
    rego.UseKey(c, func(key rego.Key, r rune) {
        switch r {
        case '+': count.Set(count.Val + 1)
        case '-': count.Set(count.Val - 1)
        }
    })
    
    // Return view
    return rego.Text(fmt.Sprintf("Count: %d", count.Val))
}
```

---

## Components with Parameters

### Method 1: Direct Parameters

Suitable for few parameters:

```go
func Badge(c rego.C, text string, color rego.Color) rego.Node {
    return rego.Text(" " + text + " ").
        Color(rego.White).
        Background(color)
}

func StatusBadge(c rego.C, status string) rego.Node {
    switch status {
    case "success":
        return Badge(c, "✓ Success", rego.Green)
    case "error":
        return Badge(c, "✗ Failed", rego.Red)
    default:
        return Badge(c, "● In Progress", rego.Yellow)
    }
}
```

### Method 2: Props Struct

Suitable for many or optional parameters:

```go
type ButtonProps struct {
    Label    string
    OnClick  func()
    Disabled bool
    Primary  bool
}

func Button(c rego.C, props ButtonProps) rego.Node {
    focused := rego.Use(c, "focused", false)
    focus := rego.UseFocus(c)
    
    rego.UseKey(c, func(key rego.Key, r rune) {
        if key == rego.KeyEnter && focus.IsFocused && !props.Disabled {
            if props.OnClick != nil {
                props.OnClick()
            }
        }
    })
    
    // Style based on state
    style := rego.Text("[" + props.Label + "]")
    
    if props.Disabled {
        return style.Dim()
    }
    if props.Primary {
        return style.Bold().Color(rego.Cyan)
    }
    if focus.IsFocused {
        return style.Color(rego.Green)
    }
    return style
}

// Usage
Button(c.Child("submit"), ButtonProps{
    Label:   "Submit",
    Primary: true,
    OnClick: func() { /* ... */ },
})
```

### Method 3: Functional Options

Suitable for flexible configuration:

```go
type inputConfig struct {
    placeholder string
    password    bool
    maxLength   int
}

type InputOption func(*inputConfig)

func WithPlaceholder(s string) InputOption {
    return func(c *inputConfig) { c.placeholder = s }
}

func WithPassword() InputOption {
    return func(c *inputConfig) { c.password = true }
}

func WithMaxLength(n int) InputOption {
    return func(c *inputConfig) { c.maxLength = n }
}

func Input(c rego.C, value string, onChange func(string), opts ...InputOption) rego.Node {
    cfg := &inputConfig{maxLength: 100}
    for _, opt := range opts {
        opt(cfg)
    }
    
    display := value
    if cfg.password {
        display = strings.Repeat("*", len(value))
    }
    if value == "" && cfg.placeholder != "" {
        display = cfg.placeholder
    }
    
    return rego.Text(display)
}

// Usage
Input(c.Child("pwd"), password, setPassword,
    WithPlaceholder("Enter password"),
    WithPassword(),
    WithMaxLength(20),
)
```

---

## Component Composition

### Basic Composition

```go
func App(c rego.C) rego.Node {
    return rego.VStack(
        Header(c.Child("header")),
        MainContent(c.Child("main")),
        Footer(c.Child("footer")),
    )
}

func Header(c rego.C) rego.Node {
    return rego.HStack(
        Logo(c.Child("logo")),
        rego.Spacer(),
        UserInfo(c.Child("user")),
    )
}
```

### Why `c.Child()`?

Each component needs its own isolated state space. `c.Child("key")` creates a **child context**:

```go
func App(c rego.C) rego.Node {
    // ❌ Wrong: Two Counters share state
    return rego.VStack(
        Counter(c),  // count = 5
        Counter(c),  // count = 5 (same one!)
    )
    
    // ✅ Correct: Each Counter has independent state
    return rego.VStack(
        Counter(c.Child("counter1")),  // count = 5
        Counter(c.Child("counter2")),  // count = 3
    )
}
```

### Components in Lists

Use `c.Child("key", index)` to create independent context for each list item:

```go
func TodoList(c rego.C) rego.Node {
    todos := rego.Use(c, "todos", []string{"Task 1", "Task 2", "Task 3"})
    
    return rego.For(todos.Val, func(todo string, i int) rego.Node {
        // Each TodoItem has its own state space
        return TodoItem(c.Child("item", i), todo)
    })
}

func TodoItem(c rego.C, text string) rego.Node {
    editing := rego.Use(c, "editing", false)
    
    if editing.Val {
        return rego.Text("[Editing] " + text).Color(rego.Yellow)
    }
    return rego.Text("• " + text)
}
```

---

## Container Components

Accept `children` parameter to wrap other nodes:

```go
// Simple container
func Card(c rego.C, title string, children ...rego.Node) rego.Node {
    return rego.VStack(
        rego.Text("┌─ " + title + " ─────────┐").Color(rego.Cyan),
        rego.VStack(children...),
        rego.Text("└───────────────────────┘").Color(rego.Cyan),
    )
}

// Usage
Card(c.Child("user-card"), "User Info",
    rego.Text("Name: John Doe"),
    rego.Text("Email: john@example.com"),
)
```

```go
// Stateful container
func Collapsible(c rego.C, title string, children ...rego.Node) rego.Node {
    expanded := rego.Use(c, "expanded", true)
    focus := rego.UseFocus(c)
    
    rego.UseKey(c, func(key rego.Key, r rune) {
        if key == rego.KeyEnter && focus.IsFocused {
            expanded.Set(!expanded.Val)
        }
    })
    
    icon := "▶"
    if expanded.Val {
        icon = "▼"
    }
    
    return rego.VStack(
        rego.Text(icon + " " + title).Bold(),
        rego.When(expanded.Val,
            rego.VStack(children...),
        ),
    )
}
```

---

## Generic Components

Use Go generics to create reusable components:

```go
// Generic select list
func SelectList[T any](
    c rego.C,
    items []T,
    renderItem func(item T, selected bool) rego.Node,
    onSelect func(item T, index int),
) rego.Node {
    selected := rego.Use(c, "selected", 0)
    
    rego.UseKey(c, func(key rego.Key, r rune) {
        switch key {
        case rego.KeyUp:
            selected.Set(max(0, selected.Val-1))
        case rego.KeyDown:
            selected.Set(min(len(items)-1, selected.Val+1))
        case rego.KeyEnter:
            if onSelect != nil && len(items) > 0 {
                onSelect(items[selected.Val], selected.Val)
            }
        }
    })
    
    return rego.For(items, func(item T, i int) rego.Node {
        return renderItem(item, i == selected.Val)
    })
}

// Usage
type File struct {
    Name string
    Size int64
}

func FileExplorer(c rego.C) rego.Node {
    files := []File{
        {"main.go", 1024},
        {"utils.go", 512},
        {"types.go", 256},
    }
    
    return SelectList(c.Child("files"), files,
        func(f File, selected bool) rego.Node {
            prefix := "  "
            if selected { prefix = "> " }
            return rego.Text(fmt.Sprintf("%s📄 %s (%d bytes)", prefix, f.Name, f.Size))
        },
        func(f File, i int) {
            fmt.Println("Selected:", f.Name)
        },
    )
}
```

---

## Focus Management

Multi-panel apps need focus management:

```go
func MultiPanelApp(c rego.C) rego.Node {
    return rego.VStack(
        LeftPanel(c.Child("left")),    // Tab order 1
        RightPanel(c.Child("right")),  // Tab order 2
        InputPanel(c.Child("input")),  // Tab order 3
    )
}

func LeftPanel(c rego.C) rego.Node {
    focus := rego.UseFocus(c)
    items := rego.Use(c, "items", []string{"A", "B", "C"})
    selected := rego.Use(c, "selected", 0)
    
    // Only handle keys when focused
    rego.UseKey(c, func(key rego.Key, r rune) {
        if !focus.IsFocused {
            return
        }
        switch key {
        case rego.KeyUp:
            selected.Set(max(0, selected.Val-1))
        case rego.KeyDown:
            selected.Set(min(len(items.Val)-1, selected.Val+1))
        }
    })
    
    // Change border color based on focus state
    borderColor := rego.Gray
    if focus.IsFocused {
        borderColor = rego.Cyan
    }
    
    return rego.Box(
        rego.For(items.Val, func(item string, i int) rego.Node {
            if i == selected.Val {
                return rego.Text("> " + item).Color(rego.Green)
            }
            return rego.Text("  " + item)
        }),
    ).Border(rego.BorderSingle).BorderColor(borderColor)
}
```

---

## Custom Hooks

Encapsulate reusable logic as custom Hooks:

### useToggle

```go
func useToggle(c rego.C, key string, initial bool) (bool, func()) {
    state := rego.Use(c, key, initial)
    toggle := func() { state.Set(!state.Val) }
    return state.Val, toggle
}

// Usage
func ExpandableSection(c rego.C, title string) rego.Node {
    expanded, toggle := useToggle(c, "expanded", false)
    focus := rego.UseFocus(c)
    
    rego.UseKey(c, func(key rego.Key, r rune) {
        if key == rego.KeyEnter && focus.IsFocused {
            toggle()
        }
    })
    
    // ...
}
```

### useSelectable

```go
func useSelectable[T any](c rego.C, items []T) (selected int, move func(delta int)) {
    state := rego.Use(c, "selected", 0)
    
    move = func(delta int) {
        newVal := state.Val + delta
        if newVal >= 0 && newVal < len(items) {
            state.Set(newVal)
        }
    }
    
    rego.UseKey(c, func(key rego.Key, r rune) {
        switch key {
        case rego.KeyUp:   move(-1)
        case rego.KeyDown: move(1)
        }
    })
    
    return state.Val, move
}
```

### useInterval

```go
func useInterval(c rego.C, callback func(), interval time.Duration) {
    callbackRef := rego.UseRef(c, &callback)
    
    rego.UseEffect(c, func() func() {
        ticker := time.NewTicker(interval)
        go func() {
            for range ticker.C {
                (*callbackRef.Current)()
                c.Refresh()
            }
        }()
        return ticker.Stop
    })
}

// Usage
func LiveClock(c rego.C) rego.Node {
    now := rego.Use(c, "now", time.Now())
    
    useInterval(c, func() {
        now.Set(time.Now())
    }, time.Second)
    
    return rego.Text(now.Val.Format("2006-01-02 15:04:05"))
}
```

### useFetch

```go
func useFetch[T any](c rego.C, url string) (data T, loading bool, err error) {
    dataState := rego.Use(c, "data", *new(T))
    loadingState := rego.Use(c, "loading", true)
    errState := rego.Use(c, "error", error(nil))
    
    rego.UseEffect(c, func() func() {
        loadingState.Set(true)
        go func() {
            resp, e := http.Get(url)
            if e != nil {
                errState.Set(e)
            } else {
                var d T
                json.NewDecoder(resp.Body).Decode(&d)
                dataState.Set(d)
            }
            loadingState.Set(false)
            c.Refresh()
        }()
        return nil
    }, url)
    
    return dataState.Val, loadingState.Val, errState.Val
}

// Usage
func UserProfile(c rego.C, userId string) rego.Node {
    user, loading, err := useFetch[User](c, "/api/users/"+userId)
    
    if loading {
        return rego.Spinner(c.Child("spin"), "Loading...")
    }
    if err != nil {
        return rego.Text("Error: " + err.Error()).Color(rego.Red)
    }
    return rego.Text("User: " + user.Name)
}
```

---

## Component Design Best Practices

### 1. Single Responsibility

```go
// ❌ Bad: One component doing too much
func TodoApp(c rego.C) rego.Node {
    // 200 lines of code...
}

// ✅ Good: Split into small components
func TodoApp(c rego.C) rego.Node {
    return rego.VStack(
        TodoHeader(c.Child("header")),
        TodoList(c.Child("list")),
        TodoInput(c.Child("input")),
        TodoFooter(c.Child("footer")),
    )
}
```

### 2. State Lifting

When multiple components need to share state, lift state to the common parent:

```go
func App(c rego.C) rego.Node {
    // State managed in parent component
    filter := rego.Use(c, "filter", "all")
    todos := rego.Use(c, "todos", []Todo{})
    
    return rego.VStack(
        // Child components only receive needed data and callbacks
        FilterBar(c.Child("filter"), filter.Val, filter.Set),
        TodoList(c.Child("list"), todos.Val, filter.Val),
    )
}

func FilterBar(c rego.C, current string, onChange func(string)) rego.Node {
    // Only responsible for UI display and event triggering
}
```

### 3. Composition Over Inheritance

```go
// ❌ Bad: Trying to "inherit"
func PrimaryButton(c rego.C, label string) rego.Node {
    // Copy all Button code...
}

// ✅ Good: Composition
func PrimaryButton(c rego.C, label string, onClick func()) rego.Node {
    return Button(c, ButtonProps{
        Label:   label,
        OnClick: onClick,
        Primary: true,
    })
}
```

### 4. Naming Conventions

| Type | Naming Rule | Example |
|------|-------------|---------|
| Component | PascalCase | `TodoItem`, `UserCard` |
| Hook | use prefix | `useToggle`, `useFetch` |
| Props | ComponentName+Props | `ButtonProps`, `CardProps` |
| Event Callback | on prefix | `onClick`, `onChange` |

---

## Complete Example: File Browser Component

```go
package main

import (
    "fmt"
    "github.com/erweixin/rego"
)

// File type
type FileEntry struct {
    Name  string
    IsDir bool
    Size  int64
}

// Props
type FileBrowserProps struct {
    Files    []FileEntry
    OnSelect func(FileEntry)
    OnOpen   func(FileEntry)
}

// Main component
func FileBrowser(c rego.C, props FileBrowserProps) rego.Node {
    selected := rego.Use(c, "selected", 0)
    focus := rego.UseFocus(c)
    
    rego.UseKey(c, func(key rego.Key, r rune) {
        if !focus.IsFocused {
            return
        }
        switch key {
        case rego.KeyUp:
            selected.Set(max(0, selected.Val-1))
        case rego.KeyDown:
            selected.Set(min(len(props.Files)-1, selected.Val+1))
        case rego.KeyEnter:
            if len(props.Files) > 0 && props.OnOpen != nil {
                props.OnOpen(props.Files[selected.Val])
            }
        }
    })
    
    borderColor := rego.Gray
    if focus.IsFocused {
        borderColor = rego.Cyan
    }
    
    return rego.Box(
        rego.VStack(
            FileBrowserHeader(c.Child("header")),
            rego.For(props.Files, func(f FileEntry, i int) rego.Node {
                return FileRow(c.Child("file", i), f, i == selected.Val)
            }),
            FileBrowserFooter(c.Child("footer"), len(props.Files)),
        ),
    ).Border(rego.BorderSingle).BorderColor(borderColor)
}

// Child component: Header
func FileBrowserHeader(c rego.C) rego.Node {
    return rego.VStack(
        rego.Text("📁 File Browser").Bold().Color(rego.Cyan),
        rego.Text("─────────────────────────────").Dim(),
    )
}

// Child component: File row
func FileRow(c rego.C, file FileEntry, selected bool) rego.Node {
    icon := "📄"
    if file.IsDir {
        icon = "📁"
    }
    
    prefix := "  "
    color := rego.White
    if selected {
        prefix = "> "
        color = rego.Green
    }
    
    text := fmt.Sprintf("%s%s %s", prefix, icon, file.Name)
    if !file.IsDir {
        text += fmt.Sprintf(" (%d bytes)", file.Size)
    }
    
    return rego.Text(text).Color(color)
}

// Child component: Footer
func FileBrowserFooter(c rego.C, count int) rego.Node {
    return rego.VStack(
        rego.Text("─────────────────────────────").Dim(),
        rego.Text(fmt.Sprintf("%d items | ↑↓ Navigate | Enter Open | Tab Switch", count)).Dim(),
    )
}

// Usage example
func App(c rego.C) rego.Node {
    files := []FileEntry{
        {"documents", true, 0},
        {"main.go", false, 1024},
        {"README.md", false, 512},
    }
    
    return FileBrowser(c.Child("browser"), FileBrowserProps{
        Files: files,
        OnOpen: func(f FileEntry) {
            fmt.Println("Opening:", f.Name)
        },
    })
}

func main() {
    rego.Run(App)
}
```

---

## Quick Reference

### Hooks (Package-level Functions)

```go
rego.Use(c, key, initial)        // State
rego.UseEffect(c, fn, deps...)   // Side effects
rego.UseKey(c, handler)          // Keyboard events
rego.UseMouse(c, handler)        // Mouse events
rego.UseMemo(c, fn, deps...)     // Memoization
rego.UseRef(c, initial)          // References
rego.UseContext(c, ctx)          // Context
rego.UseFocus(c)                 // Focus
rego.UseBridge(c, initial)       // Agent communication
```

### Context Methods

```go
c.Child(key, index...)           // Child component
c.Refresh()                      // Re-render
c.Quit()                         // Exit
c.Rect()                         // Get area
c.Wrap(node)                     // Wrap node
```

### Nodes

```go
rego.Text(s)                     // Text
rego.VStack(...) / HStack(...)   // Layout
rego.Box(child)                  // Container
rego.When(cond, node)            // Conditional rendering
rego.WhenElse(cond, a, b)        // Conditional branch
rego.For(items, fn)              // List rendering
rego.Spacer()                    // Flexible spacer
rego.Divider()                   // Divider line
rego.Center(node)                // Centering
rego.ScrollBox(c, node)          // Scroll container
rego.TailBox(c, node)            // Auto-scroll to bottom
```

### Styling

```go
.Bold().Italic().Underline().Dim().Blink()
.Color(c).Background(c)
.Width(n).Height(n).Flex(n)
.Padding(v, h).PaddingAll(t, r, b, l)
.Border(style).BorderColor(c)
.Align(rego.AlignLeft | AlignCenter | AlignRight)
.Valign(rego.AlignLeft | AlignCenter | AlignRight)
```

---

*For more examples, see the `examples/` directory.*
