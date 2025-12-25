# Rego 组件开发指南

> 如何创建和组合自定义组件

---

## 核心概念

在 Rego 中，**组件就是一个函数**：

```go
func 组件名(c rego.C, 参数...) rego.Node {
    // 使用 Hooks
    // 返回 Node
}
```

- 第一个参数必须是 `c rego.C`（组件上下文）
- 返回值是 `rego.Node`（视图节点）
- 可以有任意多个额外参数
- Hooks 使用包级函数：`rego.Use(c, ...)` 而非 `c.Use(...)`

---

## 基础组件

### 无状态组件

```go
func Divider(c rego.C) rego.Node {
    return rego.Text("────────────────────").Dim()
}

func Title(c rego.C, text string) rego.Node {
    return rego.Text(text).Bold().Color(rego.Cyan)
}
```

### 有状态组件

```go
func Counter(c rego.C) rego.Node {
    // 包级泛型函数，类型安全
    count := rego.Use(c, "count", 0)
    
    // 处理事件
    rego.UseKey(c, func(key rego.Key, r rune) {
        switch r {
        case '+': count.Set(count.Val + 1)
        case '-': count.Set(count.Val - 1)
        }
    })
    
    // 返回视图
    return rego.Text(fmt.Sprintf("Count: %d", count.Val))
}
```

---

## 带参数的组件

### 方式 1：直接传参

适合参数较少的情况：

```go
func Badge(c rego.C, text string, color rego.Color) rego.Node {
    return rego.Text(" " + text + " ").
        Color(rego.White).
        Background(color)
}

func StatusBadge(c rego.C, status string) rego.Node {
    switch status {
    case "success":
        return Badge(c, "✓ 成功", rego.Green)
    case "error":
        return Badge(c, "✗ 失败", rego.Red)
    default:
        return Badge(c, "● 进行中", rego.Yellow)
    }
}
```

### 方式 2：Props 结构体

适合参数较多或需要可选参数的情况：

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
    
    // 根据状态决定样式
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

// 使用
Button(c.Child("submit"), ButtonProps{
    Label:   "提交",
    Primary: true,
    OnClick: func() { /* ... */ },
})
```

### 方式 3：函数式选项

适合需要灵活配置的情况：

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

// 使用
Input(c.Child("pwd"), password, setPassword,
    WithPlaceholder("请输入密码"),
    WithPassword(),
    WithMaxLength(20),
)
```

---

## 组件组合

### 基本组合

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

### 为什么需要 `c.Child()`？

每个组件需要独立的状态空间。`c.Child("key")` 创建一个**子上下文**：

```go
func App(c rego.C) rego.Node {
    // ❌ 错误：两个 Counter 共享状态
    return rego.VStack(
        Counter(c),  // count = 5
        Counter(c),  // count = 5 (同一个!)
    )
    
    // ✅ 正确：每个 Counter 有独立状态
    return rego.VStack(
        Counter(c.Child("counter1")),  // count = 5
        Counter(c.Child("counter2")),  // count = 3
    )
}
```

### 列表中的组件

使用 `c.Child("key", index)` 为每个列表项创建独立上下文：

```go
func TodoList(c rego.C) rego.Node {
    todos := rego.Use(c, "todos", []string{"任务1", "任务2", "任务3"})
    
    return rego.For(todos.Val, func(todo string, i int) rego.Node {
        // 每个 TodoItem 有独立的状态空间
        return TodoItem(c.Child("item", i), todo)
    })
}

func TodoItem(c rego.C, text string) rego.Node {
    editing := rego.Use(c, "editing", false)
    
    if editing.Val {
        return rego.Text("[编辑中] " + text).Color(rego.Yellow)
    }
    return rego.Text("• " + text)
}
```

---

## 容器组件

接收 `children` 参数来包装其他节点：

```go
// 简单容器
func Card(c rego.C, title string, children ...rego.Node) rego.Node {
    return rego.VStack(
        rego.Text("┌─ " + title + " ─────────┐").Color(rego.Cyan),
        rego.VStack(children...),
        rego.Text("└───────────────────────┘").Color(rego.Cyan),
    )
}

// 使用
Card(c.Child("user-card"), "用户信息",
    rego.Text("姓名: 张三"),
    rego.Text("邮箱: zhang@example.com"),
)
```

```go
// 带状态的容器
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

## 泛型组件

利用 Go 泛型创建通用组件：

```go
// 通用选择列表
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

// 使用
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

## 焦点管理

多面板应用需要焦点管理：

```go
func MultiPanelApp(c rego.C) rego.Node {
    return rego.VStack(
        LeftPanel(c.Child("left")),    // Tab 顺序 1
        RightPanel(c.Child("right")),  // Tab 顺序 2
        InputPanel(c.Child("input")),  // Tab 顺序 3
    )
}

func LeftPanel(c rego.C) rego.Node {
    focus := rego.UseFocus(c)
    items := rego.Use(c, "items", []string{"A", "B", "C"})
    selected := rego.Use(c, "selected", 0)
    
    // 只有聚焦时才处理按键
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
    
    // 根据焦点状态改变边框颜色
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

## 自定义 Hooks

将通用逻辑封装为可复用的 Hook：

### useToggle

```go
func useToggle(c rego.C, key string, initial bool) (bool, func()) {
    state := rego.Use(c, key, initial)
    toggle := func() { state.Set(!state.Val) }
    return state.Val, toggle
}

// 使用
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

// 使用
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

// 使用
func UserProfile(c rego.C, userId string) rego.Node {
    user, loading, err := useFetch[User](c, "/api/users/"+userId)
    
    if loading {
        return rego.Spinner(c.Child("spin"), "加载中...")
    }
    if err != nil {
        return rego.Text("错误: " + err.Error()).Color(rego.Red)
    }
    return rego.Text("用户: " + user.Name)
}
```

---

## 组件设计最佳实践

### 1. 单一职责

```go
// ❌ 不好：一个组件做太多事
func TodoApp(c rego.C) rego.Node {
    // 200 行代码...
}

// ✅ 好：拆分成小组件
func TodoApp(c rego.C) rego.Node {
    return rego.VStack(
        TodoHeader(c.Child("header")),
        TodoList(c.Child("list")),
        TodoInput(c.Child("input")),
        TodoFooter(c.Child("footer")),
    )
}
```

### 2. 状态提升

当多个组件需要共享状态时，将状态提升到共同父组件：

```go
func App(c rego.C) rego.Node {
    // 状态在父组件管理
    filter := rego.Use(c, "filter", "all")
    todos := rego.Use(c, "todos", []Todo{})
    
    return rego.VStack(
        // 子组件只接收需要的数据和回调
        FilterBar(c.Child("filter"), filter.Val, filter.Set),
        TodoList(c.Child("list"), todos.Val, filter.Val),
    )
}

func FilterBar(c rego.C, current string, onChange func(string)) rego.Node {
    // 只负责 UI 展示和事件触发
}
```

### 3. 组合优于继承

```go
// ❌ 不好：试图"继承"
func PrimaryButton(c rego.C, label string) rego.Node {
    // 复制 Button 的所有代码...
}

// ✅ 好：组合
func PrimaryButton(c rego.C, label string, onClick func()) rego.Node {
    return Button(c, ButtonProps{
        Label:   label,
        OnClick: onClick,
        Primary: true,
    })
}
```

### 4. 命名约定

| 类型 | 命名规则 | 示例 |
|------|----------|------|
| 组件 | PascalCase | `TodoItem`, `UserCard` |
| Hook | use 前缀 | `useToggle`, `useFetch` |
| Props | 组件名+Props | `ButtonProps`, `CardProps` |
| 事件回调 | on 前缀 | `onClick`, `onChange` |

---

## 完整示例：文件浏览器组件

```go
package main

import (
    "fmt"
    "github.com/erweixin/rego"
)

// 文件类型
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

// 主组件
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

// 子组件：头部
func FileBrowserHeader(c rego.C) rego.Node {
    return rego.VStack(
        rego.Text("📁 File Browser").Bold().Color(rego.Cyan),
        rego.Text("─────────────────────────────").Dim(),
    )
}

// 子组件：文件行
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

// 子组件：底部
func FileBrowserFooter(c rego.C, count int) rego.Node {
    return rego.VStack(
        rego.Text("─────────────────────────────").Dim(),
        rego.Text(fmt.Sprintf("%d items | ↑↓ 移动 | Enter 打开 | Tab 切换面板", count)).Dim(),
    )
}

// 使用示例
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

## API 速查

### Hooks（包级函数）

```go
rego.Use(c, key, initial)        // 状态
rego.UseEffect(c, fn, deps...)   // 副作用
rego.UseKey(c, handler)          // 键盘事件
rego.UseMouse(c, handler)        // 鼠标事件
rego.UseMemo(c, fn, deps...)     // 缓存
rego.UseRef(c, initial)          // 引用
rego.UseContext(c, ctx)          // 上下文
rego.UseFocus(c)                 // 焦点
rego.UseBridge(c, initial)       // Agent 通信
```

### 上下文方法

```go
c.Child(key, index...)           // 子组件
c.Refresh()                      // 重渲染
c.Quit()                         // 退出
c.Rect()                         // 获取区域
c.Wrap(node)                     // 包装节点
```

### 节点

```go
rego.Text(s)                     // 文本
rego.VStack(...) / HStack(...)   // 布局
rego.Box(child)                  // 容器
rego.When(cond, node)            // 条件渲染
rego.WhenElse(cond, a, b)        // 条件分支
rego.For(items, fn)              // 列表渲染
rego.Spacer()                    // 弹性空白
rego.Divider()                   // 分隔线
rego.Center(node)                // 居中
rego.ScrollBox(c, node)          // 滚动容器
rego.TailBox(c, node)            // 自动滚动到底部
```

### 样式

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

*更多示例请参考 `examples/` 目录。*

