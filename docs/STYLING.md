# Rego Styling System Guide

Rego provides a declarative styling system based on **Chainable API**. You can build complex terminal interfaces by combining simple methods, similar to modern CSS frameworks.

---

## Core Philosophy

In Rego, styles are applied directly to nodes. Most nodes support chainable methods to modify their appearance and layout.

```go
rego.Text("Hello").
    Bold().
    Color(rego.Cyan).
    Background(rego.Black)
```

---

## 1. Text Styling

Text styles primarily apply to `rego.Text` nodes.

| Method | Description |
| :--- | :--- |
| `.Bold()` | Bold text |
| `.Italic()` | Italic text |
| `.Underline()` | Underlined text |
| `.Dim()` | Dimmed text (grayed out) |
| `.Blink()` | Blinking text |
| `.Color(rego.Color)` | Set text foreground color |
| `.Background(rego.Color)` | Set text background color |
| `.Wrap(bool)` | Enable word wrapping |

### Example
```go
rego.Text("Important Notice").Bold().Color(rego.Red)
```

---

## 2. Container Styling

Container styles primarily apply to `rego.Box` nodes.

| Method | Description |
| :--- | :--- |
| `.Border(rego.BorderStyle)` | Set border style |
| `.BorderColor(rego.Color)` | Set border color |
| `.Padding(v, h)` | Set padding (vertical, horizontal) |
| `.PaddingAll(t, r, b, l)` | Set padding individually (top, right, bottom, left) |
| `.Background(rego.Color)` | Set container background color |

### Border Styles (BorderStyle)

| Style | Appearance | Description |
| :--- | :--- | :--- |
| `rego.BorderNone` | (none) | No border (default) |
| `rego.BorderSingle` | `┌─┐` | Single line border |
| `rego.BorderDouble` | `╔═╗` | Double line border |
| `rego.BorderRounded` | `╭─╮` | Rounded corner border |
| `rego.BorderThick` | `┏━┓` | Thick line border |

### Example
```go
rego.Box(
    rego.Text("Content"),
).Border(rego.BorderRounded).BorderColor(rego.Cyan).Padding(1, 2)
```

---

## 3. Layout Properties

Layout properties control how components occupy space. They apply to `rego.Box`, `rego.VStack`, `rego.HStack`, and layout-enabled nodes.

| Method | Description |
| :--- | :--- |
| `.Width(int)` | Set fixed width (in characters) |
| `.Height(int)` | Set fixed height (in lines) |
| `.Flex(int)` | Set flex weight (proportion of remaining space in Stack) |
| `.Gap(int)` | Set spacing between child components (`VStack` and `HStack` only) |
| `.Justify(rego.Align)` | Set main axis alignment (`VStack` and `HStack` only) |
| `.Align(rego.Align)` | Set content alignment (`Box` and `Text` only) |
| `.Valign(rego.Align)` | Set vertical alignment (`Box` only) |

### Alignment Options (rego.Align)

| Value | Description |
| :--- | :--- |
| `rego.AlignLeft` | Left / Top alignment (default) |
| `rego.AlignCenter` | Center alignment |
| `rego.AlignRight` | Right / Bottom alignment |

### Vertical Alignment in Box

When a `Box` has a fixed `Height`, you can control vertical content alignment:

```go
rego.Box(content).Height(10).Valign(rego.AlignCenter)  // Vertically centered
```

### Helper Components

- **`rego.Center(node)`**: A shortcut component that automatically centers its content both horizontally and vertically within the parent container.
- **`rego.Divider()`**: A horizontal divider line that automatically fills the available width.
- **`rego.Spacer()`**: A flexible spacer that fills remaining space in a Stack.

### Example: Centered Modal Panel
```go
rego.Center(
    rego.Box(
        rego.VStack(
            rego.Text("Notice").Bold(),
            rego.Divider(),
            rego.Text("Operation completed successfully!"),
        ),
    ).Border(rego.BorderRounded).Padding(1, 2),
)
```

---

## 4. Color System

Rego provides a set of basic colors that map to standard terminal colors.

| Color | Constant |
| :--- | :--- |
| Default | `rego.Default` |
| Black | `rego.Black` |
| White | `rego.White` |
| Gray | `rego.Gray` |
| Red | `rego.Red` |
| Green | `rego.Green` |
| Blue | `rego.Blue` |
| Yellow | `rego.Yellow` |
| Cyan | `rego.Cyan` |
| Magenta | `rego.Magenta` |

---

## 5. Reusable Styles

You can create reusable `Style` objects using `rego.NewStyle()` and apply them to nodes with `.Apply()`.

```go
// Define reusable styles
var (
    TitleStyle = rego.NewStyle().
        Bold().
        Foreground(rego.Cyan).
        Padding(1, 2)

    CardStyle = rego.NewStyle().
        Border(rego.BorderRounded).
        BorderColor(rego.Gray).
        Padding(1, 2)
)

// Apply styles to nodes
rego.Text("Title").Apply(TitleStyle)
rego.Box(content).Apply(CardStyle)
```

### Available Style Methods

| Method | Description |
| :--- | :--- |
| `.Foreground(rego.Color)` | Set foreground color |
| `.Background(rego.Color)` | Set background color |
| `.Bold()` | Bold text |
| `.Italic()` | Italic text |
| `.Underline()` | Underline text |
| `.Dim()` | Dimmed text |
| `.Blink()` | Blinking text |
| `.Width(int)` | Fixed width |
| `.Height(int)` | Fixed height |
| `.Flex(int)` | Flex weight |
| `.Padding(v, h)` | Padding (vertical, horizontal) |
| `.PaddingAll(t, r, b, l)` | Padding (top, right, bottom, left) |
| `.Border(BorderStyle)` | Border style |
| `.BorderColor(Color)` | Border color |
| `.Align(Align)` | Horizontal alignment |
| `.Valign(Align)` | Vertical alignment |

---

## 6. Comprehensive Examples

### Building a Warning Panel
```go
func WarningBox(c rego.C, msg string) rego.Node {
    return rego.Box(
        rego.VStack(
            rego.Text("⚠️  WARNING").Bold().Color(rego.Yellow),
            rego.Text(msg).Wrap(true),
        ),
    ).Border(rego.BorderRounded).
       BorderColor(rego.Yellow).
       Padding(1, 2).
       Width(40)
}
```

### Responsive Layout with Flex
```go
rego.HStack(
    rego.Box(rego.Text("Sidebar")).Width(20).Border(rego.BorderSingle),
    rego.Box(rego.Text("Main Content")).Flex(1).Border(rego.BorderSingle),
)
```

### Dashboard Card
```go
func StatCard(title, value string, color rego.Color) rego.Node {
    return rego.Box(
        rego.VStack(
            rego.Text(title).Dim(),
            rego.Text(value).Bold().Color(color),
        ),
    ).Border(rego.BorderRounded).BorderColor(color).Padding(1, 2).Flex(1)
}
```

---

## 7. Word Wrapping

For longer text content, use `.Wrap(true)` to enable automatic word wrapping.

```go
rego.Text("This is a very long text that will automatically wrap based on the parent container's width, preventing overflow.").
    Wrap(true).
    Color(rego.Gray)
```

---

## 8. Dynamic Styles

Styles can be easily combined with Hooks to create interactive effects.

```go
func ToggleButton(c rego.C) rego.Node {
    active := rego.Use(c, "active", false)
    focus := rego.UseFocus(c)

    // Dynamic color based on state
    color := rego.Gray
    if active.Val {
        color = rego.Green
    }
    if focus.IsFocused {
        color = rego.Cyan
    }

    rego.UseKey(c, func(key rego.Key, r rune) {
        if key == rego.KeyEnter && focus.IsFocused {
            active.Set(!active.Val)
        }
    })

    return c.Wrap(rego.Box(
        rego.Text(rego.If(active.Val, "ON", "OFF")).Bold(),
    ).Border(rego.BorderSingle).BorderColor(color).Padding(0, 2))
}
```

### Using rego.If for Conditional Values

```go
// Ternary-like helper for inline conditionals
textColor := rego.If(isActive, rego.Green, rego.Gray)
borderStyle := rego.If(hasFocus, rego.BorderDouble, rego.BorderSingle)
```

---

## Quick Reference

### Text Node
```go
rego.Text("content").
    Bold().Italic().Underline().Dim().Blink().
    Color(rego.Cyan).
    Background(rego.Black).
    Wrap(true)
```

### Box Node
```go
rego.Box(child).
    Border(rego.BorderRounded).
    BorderColor(rego.Cyan).
    Padding(1, 2).
    Width(40).Height(10).
    Flex(1).
    Valign(rego.AlignCenter).
    Background(rego.Black)
```

### Stack Nodes
```go
rego.VStack(children...).Gap(1).Justify(rego.AlignCenter).Flex(1)
rego.HStack(children...).Gap(2).Justify(rego.AlignRight).Flex(1)
```
