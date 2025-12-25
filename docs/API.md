# Rego API Reference

> Complete API reference for the React Hooks-style Go CLI/TUI framework

---

## Table of Contents

- [Entry Function](#entry-function)
- [Component Context (C)](#component-context-c)
- [Hooks](#hooks)
  - [Use - State Management](#use---state-management)
  - [UseEffect - Side Effects](#useeffect---side-effects)
  - [UseKey - Keyboard Events](#usekey---keyboard-events)
  - [UseMouse - Mouse Events](#usemouse---mouse-events)
  - [UseFocus - Focus Management](#usefocus---focus-management)
  - [UseMemo - Memoization](#usememo---memoization)
  - [UseRef - References](#useref---references)
  - [UseContext - Cross-component Context](#usecontext---cross-component-context)
  - [UseBridge - Agent Communication](#usebridge---agent-communication)
- [Nodes](#nodes)
  - [Basic Nodes](#basic-nodes)
  - [Layout Nodes](#layout-nodes)
  - [Control Flow Nodes](#control-flow-nodes)
  - [Scroll Containers](#scroll-containers)
- [Built-in Components](#built-in-components)
- [Styling System](#styling-system)
- [Colors and Borders](#colors-and-borders)
- [Key Constants](#key-constants)
- [Mouse Event Types](#mouse-event-types)

---

## Entry Function

### Run

Start a Rego application.

```go
func Run(root func(C) Node) error
```

**Parameters**:
- `root` - Root component function

**Returns**:
- `error` - Runtime error, nil on normal exit

**Example**:

```go
func main() {
    if err := rego.Run(App); err != nil {
        log.Fatal(err)
    }
}

func App(c rego.C) rego.Node {
    return rego.Text("Hello, Rego!")
}
```

---

## Component Context (C)

`C` is the component context interface, the first parameter of every component function.

```go
type C interface {
    // Child gets a child component context
    Child(key string, index ...int) C
    
    // Refresh manually triggers a re-render
    Refresh()
    
    // Quit exits the application
    Quit()
    
    // SetCursor sets cursor position (for IME input)
    SetCursor(x, y int)
    
    // Wrap wraps a node to track its position (for mouse events)
    Wrap(node Node) *componentNode
    
    // Rect gets the component's screen area
    Rect() Rect
}
```

### Child

Creates a child component context with isolated state space.

```go
func (c C) Child(key string, index ...int) C
```

**Parameters**:
- `key` - Child component identifier
- `index` - Optional index (for list scenarios)

**Example**:

```go
func App(c rego.C) rego.Node {
    return rego.VStack(
        Header(c.Child("header")),
        Content(c.Child("content")),
        
        // Use index in lists
        rego.For(items, func(item Item, i int) rego.Node {
            return ItemRow(c.Child("item", i), item)
        }),
    )
}
```

### Refresh

Manually triggers a UI re-render.

```go
func (c C) Refresh()
```

**Use Cases**:
- After updating state in a goroutine
- After async operations complete

**Example**:

```go
rego.UseEffect(c, func() func() {
    go func() {
        data := fetchData()
        state.Set(data)
        c.Refresh()  // Refresh after async update
    }()
    return nil
})
```

### Quit

Exits the application.

```go
func (c C) Quit()
```

**Example**:

```go
rego.UseKey(c, func(key rego.Key, r rune) {
    if r == 'q' || key == rego.KeyCtrlC {
        c.Quit()
    }
})
```

### Rect

Gets the component's position and size on screen.

```go
func (c C) Rect() Rect

type Rect struct {
    X, Y, W, H int
}

func (r Rect) Contains(x, y int) bool
```

**Example**:

```go
rego.UseMouse(c, func(ev rego.MouseEvent) {
    rect := c.Rect()
    if rect.Contains(ev.X, ev.Y) {
        // Mouse is within component area
    }
})
```

---

## Hooks

### Use - State Management

Declares a state variable.

```go
func Use[T any](c C, key string, initial T) *State[T]
```

**Parameters**:
- `c` - Component context
- `key` - State identifier (unique within component)
- `initial` - Initial value

**Returns**:
- `*State[T]` - State object

**State[T] Methods**:

```go
type State[T any] struct {
    Val T  // Current value
}

// Set sets a new value and triggers re-render
func (s *State[T]) Set(value T)

// Update uses a function to update the value
func (s *State[T]) Update(fn func(old T) T)
```

**Example**:

```go
func Counter(c rego.C) rego.Node {
    // Declare different types of state
    count := rego.Use(c, "count", 0)
    name := rego.Use(c, "name", "")
    items := rego.Use(c, "items", []string{})
    
    // Read value
    fmt.Println(count.Val)
    
    // Set value
    count.Set(10)
    
    // Functional update
    count.Update(func(v int) int { return v + 1 })
    items.Update(func(list []string) []string {
        return append(list, "new item")
    })
    
    return rego.Text(fmt.Sprintf("Count: %d", count.Val))
}
```

**Use in conditionals/loops** (unlike React, Rego supports this):

```go
if showExtra {
    extra := rego.Use(c, "extra", "")  // ✅ Completely legal
}

for i := 0; i < 3; i++ {
    item := rego.Use(c, fmt.Sprintf("item-%d", i), "")  // ✅ Legal
}
```

---

### UseEffect - Side Effects

Declares a side effect with optional cleanup function.

```go
func UseEffect(c C, fn func() func(), deps ...any)
```

**Parameters**:
- `c` - Component context
- `fn` - Effect function, returns cleanup function or nil
- `deps` - Dependency list, effect re-runs when deps change

**Execution Rules**:
- Runs on first render
- When deps change: calls previous cleanup, then runs new effect
- Calls cleanup on component unmount
- With no deps: only runs on first render

**Example**:

```go
// Timer
func Timer(c rego.C) rego.Node {
    seconds := rego.Use(c, "seconds", 0)
    
    rego.UseEffect(c, func() func() {
        ticker := time.NewTicker(time.Second)
        go func() {
            for range ticker.C {
                seconds.Update(func(v int) int { return v + 1 })
                c.Refresh()
            }
        }()
        return ticker.Stop  // Cleanup function
    })  // No deps = runs once
    
    return rego.Text(fmt.Sprintf("%d seconds", seconds.Val))
}

// Data fetching (re-fetches when deps change)
func UserProfile(c rego.C, userId string) rego.Node {
    user := rego.Use(c, "user", User{})
    loading := rego.Use(c, "loading", true)
    
    rego.UseEffect(c, func() func() {
        loading.Set(true)
        go func() {
            data := fetchUser(userId)
            user.Set(data)
            loading.Set(false)
            c.Refresh()
        }()
        return nil  // No cleanup needed
    }, userId)  // Re-runs when userId changes
    
    if loading.Val {
        return rego.Spinner(c.Child("spin"), "Loading...")
    }
    return rego.Text(user.Val.Name)
}
```

---

### UseKey - Keyboard Events

Registers a keyboard event handler.

```go
func UseKey(c C, handler func(key Key, r rune))
```

**Parameters**:
- `c` - Component context
- `handler` - Event handler function
  - `key` - Special keys (arrows, enter, etc.)
  - `r` - Character (for regular key presses)

**Example**:

```go
rego.UseKey(c, func(key rego.Key, r rune) {
    // Handle special keys
    switch key {
    case rego.KeyUp:
        cursor.Update(func(v int) int { return max(0, v-1) })
    case rego.KeyDown:
        cursor.Update(func(v int) int { return v + 1 })
    case rego.KeyEnter:
        submit()
    case rego.KeyBackspace:
        deleteChar()
    }
    
    // Handle character keys
    switch r {
    case 'q':
        c.Quit()
    case '+':
        count.Set(count.Val + 1)
    }
})
```

---

### UseMouse - Mouse Events

Registers a mouse event handler.

```go
func UseMouse(c C, handler func(ev MouseEvent))
```

**MouseEvent Structure**:

```go
type MouseEvent struct {
    X, Y   int
    Button MouseButton
    Type   MouseEventType
}

type MouseEventType int
const (
    MouseEventPress MouseEventType = iota
    MouseEventRelease
    MouseEventClick
    MouseEventMove
    MouseEventScrollUp
    MouseEventScrollDown
)

type MouseButton int
const (
    MouseButtonNone MouseButton = iota
    MouseButtonLeft
    MouseButtonMiddle
    MouseButtonRight
)
```

**Example**:

```go
rego.UseMouse(c, func(ev rego.MouseEvent) {
    rect := c.Rect()
    
    switch ev.Type {
    case rego.MouseEventClick:
        if ev.Button == rego.MouseButtonLeft && rect.Contains(ev.X, ev.Y) {
            onClick()
        }
    case rego.MouseEventScrollUp:
        scrollOffset.Update(func(v int) int { return max(0, v-1) })
    case rego.MouseEventScrollDown:
        scrollOffset.Update(func(v int) int { return v + 1 })
    }
})
```

---

### UseFocus - Focus Management

Declares a component as focusable.

```go
func UseFocus(c C) FocusState
```

**FocusState Structure**:

```go
type FocusState struct {
    IsFocused bool    // Whether currently focused
    Focus     func()  // Request focus
    Blur      func()  // Release focus
}
```

**Focus Navigation**:
- `Tab` - Switch to next focusable component
- `Shift+Tab` - Switch to previous focusable component
- Mouse click auto-focuses

**Example**:

```go
func InputPanel(c rego.C) rego.Node {
    value := rego.Use(c, "value", "")
    focus := rego.UseFocus(c)
    
    rego.UseKey(c, func(key rego.Key, r rune) {
        if !focus.IsFocused {
            return  // Only handle input when focused
        }
        if r != 0 {
            value.Set(value.Val + string(r))
        }
    })
    
    // Style based on focus state
    borderColor := rego.Gray
    if focus.IsFocused {
        borderColor = rego.Cyan
    }
    
    return c.Wrap(rego.Box(
        rego.Text(value.Val),
    ).Border(rego.BorderSingle).BorderColor(borderColor))
}
```

---

### UseMemo - Memoization

Caches computed values, only recomputes when deps change.

```go
func UseMemo[T any](c C, fn func() T, deps ...any) T
```

**Parameters**:
- `c` - Component context
- `fn` - Computation function
- `deps` - Dependency list

**Returns**:
- Computed result

**Example**:

```go
func FilteredList(c rego.C, items []Item, filter string) rego.Node {
    // Only recomputes when items or filter change
    filtered := rego.UseMemo(c, func() []Item {
        result := make([]Item, 0)
        for _, item := range items {
            if strings.Contains(item.Name, filter) {
                result = append(result, item)
            }
        }
        return result
    }, items, filter)
    
    return rego.For(filtered, func(item Item, i int) rego.Node {
        return rego.Text(item.Name)
    })
}
```

---

### UseRef - References

Creates a mutable reference, solves closure traps.

```go
func UseRef[T any](c C, initial T) *Ref[T]

type Ref[T any] struct {
    Current T
}
```

**Characteristics**:
- Modifying `Ref.Current` doesn't trigger re-render
- Always accesses the latest value in closures

**Example**:

```go
func StreamingOutput(c rego.C) rego.Node {
    output := rego.Use(c, "output", "")
    
    // Ref points to the latest output pointer
    outputRef := rego.UseRef(c, &output.Val)
    
    rego.UseEffect(c, func() func() {
        go func() {
            for chunk := range stream {
                // Access latest value via Ref, avoiding stale closure
                *outputRef.Current += chunk
                c.Refresh()
            }
        }()
        return nil
    })
    
    return rego.Text(output.Val)
}
```

---

### UseContext - Cross-component Context

Reads context values provided by ancestor components.

```go
// Create context
func CreateContext[T any](name string) Context[T]

// Read context value
func UseContext[T any](c C, ctx Context[T]) T

// Provide context value (node)
func Provide[T any](ctx Context[T], value T, children ...Node) Node
```

**Example**:

```go
// Define theme context
var ThemeContext = rego.CreateContext[string]("theme")

// Root component provides value
func App(c rego.C) rego.Node {
    theme := rego.Use(c, "theme", "dark")
    
    return rego.Provide(ThemeContext, theme.Val,
        rego.VStack(
            Header(c.Child("header")),
            Content(c.Child("content")),
        ),
    )
}

// Child component consumes value
func ThemedButton(c rego.C, text string) rego.Node {
    theme := rego.UseContext(c, ThemeContext)
    
    if theme == "dark" {
        return rego.Text(text).Color(rego.White).Background(rego.Black)
    }
    return rego.Text(text).Color(rego.Black).Background(rego.White)
}
```

---

### UseBridge - Agent Communication

Creates a bidirectional communication bridge between UI and background Agent.

```go
func UseBridge[S any, Q any, A any](c C, initial S) *Bridge[S, Q, A]
```

**Type Parameters**:
- `S` - State type
- `Q` - Question type (Agent requests from UI)
- `A` - Answer type (UI replies to Agent)

**Bridge Methods**:

```go
type Bridge[S, Q, A any] struct { ... }

// State gets current state
func (b *Bridge[S, Q, A]) State() S

// HasInteraction checks for pending interaction request
func (b *Bridge[S, Q, A]) HasInteraction() bool

// Interaction gets current interaction request
func (b *Bridge[S, Q, A]) Interaction() Q

// Submit submits answer, unblocking Agent
func (b *Bridge[S, Q, A]) Submit(answer A)

// Handle gets the Agent-side handle
func (b *Bridge[S, Q, A]) Handle() Handle[S, Q, A]
```

**Handle Interface** (Agent-side):

```go
type Handle[S, Q, A any] interface {
    Update(state S)       // Update state (non-blocking)
    Ask(question Q) A     // Request user interaction (blocking)
}
```

**Example**:

```go
type AgentState struct {
    Response string
    Status   string
}

type Confirm struct {
    Message string
}

func AgentUI(c rego.C) rego.Node {
    bridge := rego.UseBridge[AgentState, Confirm, bool](c, AgentState{})
    
    rego.UseEffect(c, func() func() {
        go func() {
            handle := bridge.Handle()
            
            // Streaming updates
            for token := range stream {
                handle.Update(AgentState{
                    Response: current + token,
                    Status:   "streaming",
                })
            }
            
            // Request user confirmation (blocking)
            confirmed := handle.Ask(Confirm{Message: "Apply changes?"})
            if confirmed {
                applyChanges()
            }
        }()
        return nil
    })
    
    return rego.VStack(
        rego.Markdown(bridge.State().Response),
        
        rego.When(bridge.HasInteraction(),
            rego.HStack(
                rego.Text(bridge.Interaction().Message),
                rego.Button(c.Child("yes"), rego.ButtonProps{
                    Label: "Yes",
                    OnClick: func() { bridge.Submit(true) },
                }),
                rego.Button(c.Child("no"), rego.ButtonProps{
                    Label: "No",
                    OnClick: func() { bridge.Submit(false) },
                }),
            ),
        ),
    )
}
```

---

## Nodes

### Basic Nodes

#### Text

Creates a text node.

```go
func Text(content string) *textNode
```

**Style Methods**:

```go
text := rego.Text("Hello").
    Bold().           // Bold
    Italic().         // Italic
    Underline().      // Underline
    Dim().            // Dimmed
    Blink().          // Blinking
    Color(rego.Cyan).        // Foreground color
    Background(rego.Black).  // Background color
    Wrap(true)        // Word wrap
```

#### Empty

Creates an empty node that takes no space.

```go
func Empty() *emptyNode
```

#### Spacer

Creates a flexible spacer that fills remaining space.

```go
func Spacer() *spacerNode
```

**Example**:

```go
rego.HStack(
    rego.Text("Left"),
    rego.Spacer(),      // Fills middle space
    rego.Text("Right"),
)
```

#### Divider

Creates a horizontal divider line.

```go
func Divider() *dividerNode

divider := rego.Divider().
    Char('═').          // Set divider character
    Color(rego.Gray)    // Set color
```

#### Cursor

Marks cursor position (for IME input positioning).

```go
func Cursor(c C) Node
```

---

### Layout Nodes

#### VStack

Vertical stack layout.

```go
func VStack(children ...Node) *vstackNode
```

**Style Methods**:

```go
rego.VStack(children...).
    Gap(1).                    // Gap between children
    Justify(rego.AlignCenter). // Main axis alignment (Top/Center/Bottom)
    Padding(1, 2).             // Padding (vertical, horizontal)
    Height(10).                // Fixed height
    Flex(1).                   // Flex weight
    Background(rego.Black)     // Background color
```

#### HStack

Horizontal stack layout.

```go
func HStack(children ...Node) *hstackNode
```

**Style Methods**:

```go
rego.HStack(children...).
    Gap(2).                   // Gap between children
    Justify(rego.AlignRight). // Main axis alignment (Left/Center/Right)
    Padding(0, 1).            // Padding
    Height(1).                // Fixed height
    Flex(1)                   // Flex weight
```

#### Box

Container node with border and padding support.

```go
func Box(child Node) *boxNode
```

**Style Methods**:

```go
rego.Box(child).
    Border(rego.BorderRounded).   // Border style
    BorderColor(rego.Cyan).       // Border color
    Padding(1, 2).                // Padding (vertical, horizontal)
    PaddingAll(1, 2, 1, 2).       // Padding (top, right, bottom, left)
    Width(40).                    // Fixed width
    Height(10).                   // Fixed height
    Flex(1).                      // Flex weight
    Valign(rego.AlignCenter).     // Vertical alignment
    Background(rego.Black)        // Background color
```

#### Center

Centering helper component.

```go
func Center(child Node) Node
```

**Example**:

```go
rego.Center(
    rego.Box(
        rego.Text("Centered Content"),
    ).Border(rego.BorderSingle),
)
```

---

### Control Flow Nodes

#### When

Conditional rendering.

```go
func When(condition bool, node Node) *whenNode
```

**Example**:

```go
rego.When(loading, rego.Spinner(c.Child("spin"), "Loading..."))
rego.When(error != nil, rego.Text(error.Error()).Color(rego.Red))
```

#### WhenElse

Conditional rendering with else branch.

```go
func WhenElse(condition bool, trueNode, falseNode Node) *whenElseNode
```

**Example**:

```go
rego.WhenElse(loggedIn,
    UserPanel(c.Child("user")),
    LoginForm(c.Child("login")),
)
```

#### For

List rendering.

```go
func For[T any](items []T, render func(item T, index int) Node) Node
```

**Example**:

```go
rego.For(todos, func(todo Todo, i int) rego.Node {
    return TodoItem(c.Child("todo", i), todo, i == selected)
})
```

---

### Scroll Containers

#### ScrollBox

Scrollable container.

```go
func ScrollBox(c C, child Node) *componentNode
```

**Features**:
- Mouse wheel support
- Auto-displays scrollbar

#### TailBox

Auto-scroll to bottom container (ideal for logs/chat).

```go
func TailBox(c C, child Node) *componentNode
```

**Features**:
- New content auto-scrolls to bottom
- Manual scroll up pauses auto-scroll

---

## Built-in Components

### Button

Button component.

```go
type ButtonProps struct {
    Label   string   // Button text
    OnClick func()   // Click callback
    Primary bool     // Whether primary button
}

func Button(c C, props ButtonProps) Node
```

### TextInput

Text input component.

```go
type TextInputProps struct {
    Value       string         // Current value
    Placeholder string         // Placeholder
    Label       string         // Label
    Width       int            // Width
    Height      int            // Height (for multiline)
    Multiline   bool           // Whether multiline
    Password    bool           // Whether password mode
    OnChanged   func(string)   // Value change callback
    OnSubmit    func(string)   // Enter submit callback
}

func TextInput(c C, props TextInputProps) Node
```

### Checkbox

Checkbox component.

```go
type CheckboxProps struct {
    Label     string       // Label
    Checked   bool         // Whether checked
    OnChanged func(bool)   // State change callback
}

func Checkbox(c C, props CheckboxProps) Node
```

### Spinner

Loading animation component.

```go
func Spinner(c C, label string) Node
```

### Markdown

Markdown rendering component.

```go
func Markdown(content string) *markdownNode

md := rego.Markdown(content).
    Theme("dark")  // Theme: "dark", "light", "notty"
```

---

## Styling System

### Style Object

Reusable style object.

```go
style := rego.NewStyle().
    Foreground(rego.White).
    Background(rego.Black).
    Bold().
    Padding(1, 2).
    Border(rego.BorderSingle)

rego.Text("Hello").Apply(style)
rego.Box(child).Apply(style)
```

### Alignment

```go
type Align int

const (
    AlignLeft   Align = iota  // Left / Top alignment
    AlignCenter               // Center alignment
    AlignRight                // Right / Bottom alignment
)
```

---

## Colors and Borders

### Color Constants

```go
const (
    Default Color = iota  // Terminal default
    Black
    Red
    Green
    Yellow
    Blue
    Magenta
    Cyan
    White
    Gray
)
```

### Border Styles

```go
const (
    BorderNone    BorderStyle = iota  // No border
    BorderSingle                       // ┌─┐
    BorderDouble                       // ╔═╗
    BorderRounded                      // ╭─╮
    BorderThick                        // ┏━┓
)
```

---

## Key Constants

```go
const (
    KeyNone Key = iota
    KeyUp
    KeyDown
    KeyLeft
    KeyRight
    KeyEnter
    KeyEscape
    KeyTab
    KeyBackspace
    KeyDelete
    KeyHome
    KeyEnd
    KeyPgUp
    KeyPgDown
    KeyInsert
    KeyF1
    KeyF2
    // ... F3-F12
    KeyCtrlA
    KeyCtrlB
    KeyCtrlC
    // ... Ctrl+D-Z
)
```

---

## Mouse Event Types

```go
type MouseEventType int

const (
    MouseEventPress      MouseEventType = iota  // Press
    MouseEventRelease                            // Release
    MouseEventClick                              // Click
    MouseEventMove                               // Move
    MouseEventScrollUp                           // Scroll up
    MouseEventScrollDown                         // Scroll down
)

type MouseButton int

const (
    MouseButtonNone   MouseButton = iota  // No button
    MouseButtonLeft                        // Left button
    MouseButtonMiddle                      // Middle button
    MouseButtonRight                       // Right button
)
```

---

## Testing Utilities

Rego provides a testing package `rego/testing`:

```go
import regotest "github.com/erweixin/rego/testing"

func TestMyComponent(t *testing.T) {
    // Create mock screen
    screen := regotest.NewMockScreen(80, 24)
    
    // Create test runtime
    rt := regotest.NewTestRuntime(MyComponent, 80, 24)
    
    // Execute render
    rt.Render()
    
    // Get screen content
    content := rt.Screen.GetContentString()
    
    // Simulate key press
    rt.DispatchKey(rego.KeyEnter, 0, 0)
    
    // Assert results
    if !strings.Contains(content, "expected text") {
        t.Error("Expected text not found")
    }
}
```

---

## Quick Reference

### Hooks

| Hook | Signature | Description |
|------|-----------|-------------|
| `Use` | `Use[T](c, key, initial) *State[T]` | State management |
| `UseEffect` | `UseEffect(c, fn, deps...)` | Side effects |
| `UseKey` | `UseKey(c, handler)` | Keyboard events |
| `UseMouse` | `UseMouse(c, handler)` | Mouse events |
| `UseFocus` | `UseFocus(c) FocusState` | Focus management |
| `UseMemo` | `UseMemo[T](c, fn, deps...) T` | Memoization |
| `UseRef` | `UseRef[T](c, initial) *Ref[T]` | References |
| `UseContext` | `UseContext[T](c, ctx) T` | Context consumption |
| `UseBridge` | `UseBridge[S,Q,A](c, init) *Bridge` | Agent communication |

### Nodes

| Category | APIs |
|----------|------|
| **Basic** | `Text`, `Empty`, `Spacer`, `Divider`, `Cursor` |
| **Layout** | `VStack`, `HStack`, `Box`, `Center` |
| **Control** | `When`, `WhenElse`, `For` |
| **Scroll** | `ScrollBox`, `TailBox` |
| **Components** | `Button`, `TextInput`, `Checkbox`, `Spinner`, `Markdown` |

### Context Methods

| Method | Description |
|--------|-------------|
| `c.Child(key, i...)` | Get child component context |
| `c.Refresh()` | Trigger re-render |
| `c.Quit()` | Exit application |
| `c.Rect()` | Get component area |
| `c.Wrap(node)` | Wrap node |

---

*This documentation is based on Rego v0.1.0-beta*

