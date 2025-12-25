# Rego 样式系统指南

Rego 提供了一套基于**链式调用 (Chainable API)** 的声明式样式系统。你可以像在现代 CSS 框架中一样，通过组合简单的方法来构建复杂的终端界面。

---

## 核心理念

在 Rego 中，样式是直接作用在节点 (Node) 上的。大多数节点都支持链式调用来修改其外观和布局。

```go
rego.Text("Hello").
    Bold().
    Color(rego.Cyan).
    Background(rego.Black)
```

---

## 1. 文字样式 (Text Styling)

文字样式主要作用于 `rego.Text` 节点。

| 方法 | 说明 |
| :--- | :--- |
| `.Bold()` | 加粗文字 |
| `.Italic()` | 斜体文字 |
| `.Underline()` | 下划线 |
| `.Dim()` | 调暗文字（变灰） |
| `.Blink()` | 文字闪烁 |
| `.Color(rego.Color)` | 设置文字前景色 |
| `.Background(rego.Color)` | 设置文字背景色 |
| `.Wrap(bool)` | 启用自动换行 |

### 示例
```go
rego.Text("重要提示").Bold().Color(rego.Red)
```

---

## 2. 容器样式 (Container Styling)

容器样式主要作用于 `rego.Box` 节点。

| 方法 | 说明 |
| :--- | :--- |
| `.Border(rego.BorderStyle)` | 设置边框样式 |
| `.BorderColor(rego.Color)` | 设置边框颜色 |
| `.Padding(v, h)` | 设置内边距（垂直, 水平） |
| `.PaddingAll(t, r, b, l)` | 分别设置上、右、下、左内边距 |
| `.Background(rego.Color)` | 设置整个容器的背景色 |

### 边框样式 (BorderStyle)

| 样式 | 外观 | 说明 |
| :--- | :--- | :--- |
| `rego.BorderNone` | (无) | 无边框（默认） |
| `rego.BorderSingle` | `┌─┐` | 单线边框 |
| `rego.BorderDouble` | `╔═╗` | 双线边框 |
| `rego.BorderRounded` | `╭─╮` | 圆角边框 |
| `rego.BorderThick` | `┏━┓` | 粗线边框 |

### 示例
```go
rego.Box(
    rego.Text("内容"),
).Border(rego.BorderRounded).BorderColor(rego.Cyan).Padding(1, 2)
```

---

## 3. 布局属性 (Layout Properties)

布局属性用于控制组件在空间中的占据方式，适用于 `rego.Box`、`rego.VStack`、`rego.HStack` 以及支持布局的节点。

| 方法 | 说明 |
| :--- | :--- |
| `.Width(int)` | 设置固定宽度（字符数） |
| `.Height(int)` | 设置固定高度（行数） |
| `.Flex(int)` | 设置弹性权重（在 Stack 中占据剩余空间的比例） |
| `.Gap(int)` | 设置子组件之间的间距（仅限 `VStack` 和 `HStack`） |
| `.Justify(rego.Align)` | 设置主轴对齐方式（仅限 `VStack` 和 `HStack`） |
| `.Align(rego.Align)` | 设置内容对齐方式（仅限 `Box` 和 `Text`） |
| `.Valign(rego.Align)` | 设置垂直对齐方式（仅限 `Box`） |

### 对齐选项 (rego.Align)

| 值 | 说明 |
| :--- | :--- |
| `rego.AlignLeft` | 左对齐 / 顶部对齐（默认） |
| `rego.AlignCenter` | 居中对齐 |
| `rego.AlignRight` | 右对齐 / 底部对齐 |

### Box 中的垂直对齐

当 `Box` 设置了固定 `Height` 时，可以控制内容的垂直对齐方式：

```go
rego.Box(content).Height(10).Valign(rego.AlignCenter)  // 垂直居中
```

### 辅助组件

- **`rego.Center(node)`**: 快捷组件，会自动在父容器中水平和垂直双向居中其内容。
- **`rego.Divider()`**: 自动撑满宽度的水平分隔线。
- **`rego.Spacer()`**: 弹性空白，填充 Stack 中的剩余空间。

### 示例：创建一个屏幕正中心的居中面板
```go
rego.Center(
    rego.Box(
        rego.VStack(
            rego.Text("提示").Bold(),
            rego.Divider(),
            rego.Text("操作已成功完成！"),
        ),
    ).Border(rego.BorderRounded).Padding(1, 2),
)
```

---

## 4. 颜色系统 (Color System)

Rego 内置了一组基础颜色，映射到终端的标准颜色。

| 颜色 | 常量 |
| :--- | :--- |
| 默认色 | `rego.Default` |
| 黑色 | `rego.Black` |
| 白色 | `rego.White` |
| 灰色 | `rego.Gray` |
| 红色 | `rego.Red` |
| 绿色 | `rego.Green` |
| 蓝色 | `rego.Blue` |
| 黄色 | `rego.Yellow` |
| 青色 | `rego.Cyan` |
| 品红 | `rego.Magenta` |

---

## 5. 可复用样式 (Reusable Styles)

你可以使用 `rego.NewStyle()` 创建可复用的 `Style` 对象，并通过 `.Apply()` 方法应用到节点上。

```go
// 定义可复用样式
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

// 应用样式到节点
rego.Text("标题").Apply(TitleStyle)
rego.Box(content).Apply(CardStyle)
```

### 可用的 Style 方法

| 方法 | 说明 |
| :--- | :--- |
| `.Foreground(rego.Color)` | 设置前景色 |
| `.Background(rego.Color)` | 设置背景色 |
| `.Bold()` | 加粗 |
| `.Italic()` | 斜体 |
| `.Underline()` | 下划线 |
| `.Dim()` | 调暗 |
| `.Blink()` | 闪烁 |
| `.Width(int)` | 固定宽度 |
| `.Height(int)` | 固定高度 |
| `.Flex(int)` | 弹性权重 |
| `.Padding(v, h)` | 内边距（垂直，水平） |
| `.PaddingAll(t, r, b, l)` | 内边距（上、右、下、左） |
| `.Border(BorderStyle)` | 边框样式 |
| `.BorderColor(Color)` | 边框颜色 |
| `.Align(Align)` | 水平对齐 |
| `.Valign(Align)` | 垂直对齐 |

---

## 6. 综合示例

### 构建一个警告面板
```go
func WarningBox(c rego.C, msg string) rego.Node {
    return rego.Box(
        rego.VStack(
            rego.Text("⚠️  警告").Bold().Color(rego.Yellow),
            rego.Text(msg).Wrap(true),
        ),
    ).Border(rego.BorderRounded).
       BorderColor(rego.Yellow).
       Padding(1, 2).
       Width(40)
}
```

### 响应式布局与 Flex
```go
rego.HStack(
    rego.Box(rego.Text("侧边栏")).Width(20).Border(rego.BorderSingle),
    rego.Box(rego.Text("主内容")).Flex(1).Border(rego.BorderSingle),
)
```

### 仪表盘卡片
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

## 7. 自动换行 (Word Wrap)

对于较长的文本，可以使用 `.Wrap(true)` 开启自动换行。

```go
rego.Text("这是一段非常长的文本，它会自动根据父容器的宽度进行换行，而不会溢出屏幕。").
    Wrap(true).
    Color(rego.Gray)
```

---

## 8. 动态样式

样式可以非常容易地与 Hooks 结合，实现交互式效果。

```go
func ToggleButton(c rego.C) rego.Node {
    active := rego.Use(c, "active", false)
    focus := rego.UseFocus(c)

    // 根据状态动态设置颜色
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
        rego.Text(rego.If(active.Val, "开", "关")).Bold(),
    ).Border(rego.BorderSingle).BorderColor(color).Padding(0, 2))
}
```

### 使用 rego.If 进行条件取值

```go
// 类似三元运算符的辅助函数
textColor := rego.If(isActive, rego.Green, rego.Gray)
borderStyle := rego.If(hasFocus, rego.BorderDouble, rego.BorderSingle)
```

---

## 速查表

### Text 节点
```go
rego.Text("内容").
    Bold().Italic().Underline().Dim().Blink().
    Color(rego.Cyan).
    Background(rego.Black).
    Wrap(true)
```

### Box 节点
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

### Stack 节点
```go
rego.VStack(children...).Gap(1).Justify(rego.AlignCenter).Flex(1)
rego.HStack(children...).Gap(2).Justify(rego.AlignRight).Flex(1)
```

