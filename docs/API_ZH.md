# Rego API 参考文档

> React Hooks 风格的 Go CLI/TUI 框架完整 API 参考

---

## 目录

- [入口函数](#入口函数)
- [组件上下文 (C)](#组件上下文-c)
- [Hooks](#hooks)
  - [Use - 状态管理](#use---状态管理)
  - [UseEffect - 副作用](#useeffect---副作用)
  - [UseKey - 键盘事件](#usekey---键盘事件)
  - [UseMouse - 鼠标事件](#usemouse---鼠标事件)
  - [UseFocus - 焦点管理](#usefocus---焦点管理)
  - [UseMemo - 缓存计算](#usememo---缓存计算)
  - [UseRef - 引用](#useref---引用)
  - [UseContext - 跨组件上下文](#usecontext---跨组件上下文)
  - [UseBridge - Agent 通信](#usebridge---agent-通信)
- [节点 (Node)](#节点-node)
  - [基础节点](#基础节点)
  - [布局节点](#布局节点)
  - [控制流节点](#控制流节点)
  - [滚动容器](#滚动容器)
- [内置组件](#内置组件)
- [样式系统](#样式系统)
- [颜色与边框](#颜色与边框)
- [按键常量](#按键常量)
- [鼠标事件](#鼠标事件类型)

---

## 入口函数

### Run

启动 Rego 应用。

```go
func Run(root func(C) Node) error
```

**参数**:
- `root` - 根组件函数

**返回值**:
- `error` - 运行时错误，正常退出返回 nil

**示例**:

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

## 组件上下文 (C)

`C` 是组件上下文接口，每个组件函数的第一个参数。

```go
type C interface {
    // Child 获取子组件上下文
    Child(key string, index ...int) C
    
    // Refresh 手动触发重渲染
    Refresh()
    
    // Quit 退出应用
    Quit()
    
    // SetCursor 设置光标位置（用于 IME 输入）
    SetCursor(x, y int)
    
    // Wrap 包装节点以追踪位置（用于鼠标事件）
    Wrap(node Node) *componentNode
    
    // Rect 获取当前组件的屏幕区域
    Rect() Rect
}
```

### Child

创建子组件上下文，每个子组件拥有独立的状态空间。

```go
func (c C) Child(key string, index ...int) C
```

**参数**:
- `key` - 子组件标识符
- `index` - 可选的索引（用于列表场景）

**示例**:

```go
func App(c rego.C) rego.Node {
    return rego.VStack(
        Header(c.Child("header")),
        Content(c.Child("content")),
        
        // 列表中使用 index
        rego.For(items, func(item Item, i int) rego.Node {
            return ItemRow(c.Child("item", i), item)
        }),
    )
}
```

### Refresh

手动触发界面重渲染。

```go
func (c C) Refresh()
```

**使用场景**:
- 在 goroutine 中更新状态后需要刷新界面
- 异步操作完成后刷新

**示例**:

```go
rego.UseEffect(c, func() func() {
    go func() {
        data := fetchData()
        state.Set(data)
        c.Refresh()  // 异步更新后刷新
    }()
    return nil
})
```

### Quit

退出应用程序。

```go
func (c C) Quit()
```

**示例**:

```go
rego.UseKey(c, func(key rego.Key, r rune) {
    if r == 'q' || key == rego.KeyCtrlC {
        c.Quit()
    }
})
```

### Rect

获取当前组件在屏幕上的位置和大小。

```go
func (c C) Rect() Rect

type Rect struct {
    X, Y, W, H int
}

func (r Rect) Contains(x, y int) bool
```

**示例**:

```go
rego.UseMouse(c, func(ev rego.MouseEvent) {
    rect := c.Rect()
    if rect.Contains(ev.X, ev.Y) {
        // 鼠标在组件区域内
    }
})
```

---

## Hooks

### Use - 状态管理

声明一个状态变量。

```go
func Use[T any](c C, key string, initial T) *State[T]
```

**参数**:
- `c` - 组件上下文
- `key` - 状态标识符（同一组件内唯一）
- `initial` - 初始值

**返回值**:
- `*State[T]` - 状态对象

**State[T] 方法**:

```go
type State[T any] struct {
    Val T  // 当前值
}

// Set 设置新值并触发重渲染
func (s *State[T]) Set(value T)

// Update 使用函数更新值
func (s *State[T]) Update(fn func(old T) T)
```

**示例**:

```go
func Counter(c rego.C) rego.Node {
    // 声明不同类型的状态
    count := rego.Use(c, "count", 0)
    name := rego.Use(c, "name", "")
    items := rego.Use(c, "items", []string{})
    
    // 读取值
    fmt.Println(count.Val)
    
    // 设置值
    count.Set(10)
    
    // 函数式更新
    count.Update(func(v int) int { return v + 1 })
    items.Update(func(list []string) []string {
        return append(list, "new item")
    })
    
    return rego.Text(fmt.Sprintf("Count: %d", count.Val))
}
```

**在条件/循环中使用**（与 React 不同，Rego 支持）:

```go
if showExtra {
    extra := rego.Use(c, "extra", "")  // ✅ 完全合法
}

for i := 0; i < 3; i++ {
    item := rego.Use(c, fmt.Sprintf("item-%d", i), "")  // ✅ 合法
}
```

---

### UseEffect - 副作用

声明副作用，返回可选的清理函数。

```go
func UseEffect(c C, fn func() func(), deps ...any)
```

**参数**:
- `c` - 组件上下文
- `fn` - 副作用函数，返回清理函数或 nil
- `deps` - 依赖列表，依赖变化时重新执行

**执行规则**:
- 首次渲染时执行
- 依赖项变化时，先调用上一次的清理函数，再执行新的副作用
- 组件卸载时调用清理函数
- 无依赖参数时，只在首次渲染执行

**示例**:

```go
// 定时器
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
        return ticker.Stop  // 清理函数
    })  // 无依赖 = 只执行一次
    
    return rego.Text(fmt.Sprintf("%d seconds", seconds.Val))
}

// 数据获取（依赖变化时重新获取）
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
        return nil  // 无需清理
    }, userId)  // userId 变化时重新执行
    
    if loading.Val {
        return rego.Spinner(c.Child("spin"), "Loading...")
    }
    return rego.Text(user.Val.Name)
}
```

---

### UseKey - 键盘事件

注册键盘事件处理器。

```go
func UseKey(c C, handler func(key Key, r rune))
```

**参数**:
- `c` - 组件上下文
- `handler` - 事件处理函数
  - `key` - 特殊按键（如方向键、回车等）
  - `r` - 字符（普通按键时有值）

**示例**:

```go
rego.UseKey(c, func(key rego.Key, r rune) {
    // 处理特殊按键
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
    
    // 处理字符按键
    switch r {
    case 'q':
        c.Quit()
    case '+':
        count.Set(count.Val + 1)
    }
})
```

---

### UseMouse - 鼠标事件

注册鼠标事件处理器。

```go
func UseMouse(c C, handler func(ev MouseEvent))
```

**MouseEvent 结构**:

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

**示例**:

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

### UseFocus - 焦点管理

声明组件可获得焦点。

```go
func UseFocus(c C) FocusState
```

**FocusState 结构**:

```go
type FocusState struct {
    IsFocused bool    // 当前是否有焦点
    Focus     func()  // 请求焦点
    Blur      func()  // 放弃焦点
}
```

**焦点导航**:
- `Tab` - 切换到下一个可聚焦组件
- `Shift+Tab` - 切换到上一个可聚焦组件
- 鼠标点击自动聚焦

**示例**:

```go
func InputPanel(c rego.C) rego.Node {
    value := rego.Use(c, "value", "")
    focus := rego.UseFocus(c)
    
    rego.UseKey(c, func(key rego.Key, r rune) {
        if !focus.IsFocused {
            return  // 只有获得焦点时才处理输入
        }
        if r != 0 {
            value.Set(value.Val + string(r))
        }
    })
    
    // 根据焦点状态设置样式
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

### UseMemo - 缓存计算

缓存计算结果，只在依赖变化时重新计算。

```go
func UseMemo[T any](c C, fn func() T, deps ...any) T
```

**参数**:
- `c` - 组件上下文
- `fn` - 计算函数
- `deps` - 依赖列表

**返回值**:
- 计算结果

**示例**:

```go
func FilteredList(c rego.C, items []Item, filter string) rego.Node {
    // 只在 items 或 filter 变化时重新过滤
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

### UseRef - 引用

创建一个可变引用，解决闭包陷阱。

```go
func UseRef[T any](c C, initial T) *Ref[T]

type Ref[T any] struct {
    Current T
}
```

**特点**:
- `Ref.Current` 的修改不会触发重渲染
- 在闭包中始终访问最新值

**示例**:

```go
func StreamingOutput(c rego.C) rego.Node {
    output := rego.Use(c, "output", "")
    
    // Ref 指向最新的 output 指针
    outputRef := rego.UseRef(c, &output.Val)
    
    rego.UseEffect(c, func() func() {
        go func() {
            for chunk := range stream {
                // 通过 Ref 访问最新值，避免闭包捕获旧值
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

### UseContext - 跨组件上下文

读取祖先组件提供的上下文值。

```go
// 创建上下文
func CreateContext[T any](name string) Context[T]

// 读取上下文值
func UseContext[T any](c C, ctx Context[T]) T

// 提供上下文值（节点）
func Provide[T any](ctx Context[T], value T, children ...Node) Node
```

**示例**:

```go
// 定义主题上下文
var ThemeContext = rego.CreateContext[string]("theme")

// 根组件提供值
func App(c rego.C) rego.Node {
    theme := rego.Use(c, "theme", "dark")
    
    return rego.Provide(ThemeContext, theme.Val,
        rego.VStack(
            Header(c.Child("header")),
            Content(c.Child("content")),
        ),
    )
}

// 子组件消费值
func ThemedButton(c rego.C, text string) rego.Node {
    theme := rego.UseContext(c, ThemeContext)
    
    if theme == "dark" {
        return rego.Text(text).Color(rego.White).Background(rego.Black)
    }
    return rego.Text(text).Color(rego.Black).Background(rego.White)
}
```

---

### UseBridge - Agent 通信

创建 UI 与后台 Agent 的双向通信桥梁。

```go
func UseBridge[S any, Q any, A any](c C, initial S) *Bridge[S, Q, A]
```

**类型参数**:
- `S` - 状态类型
- `Q` - 问题类型（Agent 向 UI 请求）
- `A` - 答案类型（UI 回复 Agent）

**Bridge 方法**:

```go
type Bridge[S, Q, A any] struct { ... }

// State 获取当前状态
func (b *Bridge[S, Q, A]) State() S

// HasInteraction 检查是否有挂起的交互请求
func (b *Bridge[S, Q, A]) HasInteraction() bool

// Interaction 获取当前交互请求
func (b *Bridge[S, Q, A]) Interaction() Q

// Submit 提交答案，解除 Agent 阻塞
func (b *Bridge[S, Q, A]) Submit(answer A)

// Handle 获取 Agent 侧的句柄
func (b *Bridge[S, Q, A]) Handle() Handle[S, Q, A]
```

**Handle 接口**（Agent 侧使用）:

```go
type Handle[S, Q, A any] interface {
    Update(state S)       // 更新状态（非阻塞）
    Ask(question Q) A     // 请求用户交互（阻塞）
}
```

**示例**:

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
            
            // 流式更新
            for token := range stream {
                handle.Update(AgentState{
                    Response: current + token,
                    Status:   "streaming",
                })
            }
            
            // 请求用户确认（阻塞）
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

## 节点 (Node)

### 基础节点

#### Text

创建文本节点。

```go
func Text(content string) *textNode
```

**样式方法**:

```go
text := rego.Text("Hello").
    Bold().           // 粗体
    Italic().         // 斜体
    Underline().      // 下划线
    Dim().            // 暗色
    Blink().          // 闪烁
    Color(rego.Cyan).        // 前景色
    Background(rego.Black).  // 背景色
    Wrap(true)        // 自动换行
```

#### Empty

创建空节点，不占用空间。

```go
func Empty() *emptyNode
```

#### Spacer

创建弹性空白节点，自动填充剩余空间。

```go
func Spacer() *spacerNode
```

**示例**:

```go
rego.HStack(
    rego.Text("Left"),
    rego.Spacer(),      // 填充中间空间
    rego.Text("Right"),
)
```

#### Divider

创建水平分隔线。

```go
func Divider() *dividerNode

divider := rego.Divider().
    Char('═').          // 设置分隔字符
    Color(rego.Gray)    // 设置颜色
```

#### Cursor

标记光标位置（用于 IME 输入定位）。

```go
func Cursor(c C) Node
```

---

### 布局节点

#### VStack

垂直堆叠布局。

```go
func VStack(children ...Node) *vstackNode
```

**样式方法**:

```go
rego.VStack(children...).
    Gap(1).                    // 子元素间距
    Justify(rego.AlignCenter). // 主轴对齐（Top/Center/Bottom）
    Padding(1, 2).             // 内边距（垂直, 水平）
    Height(10).                // 固定高度
    Flex(1).                   // 弹性权重
    Background(rego.Black)     // 背景色
```

#### HStack

水平排列布局。

```go
func HStack(children ...Node) *hstackNode
```

**样式方法**:

```go
rego.HStack(children...).
    Gap(2).                   // 子元素间距
    Justify(rego.AlignRight). // 主轴对齐（Left/Center/Right）
    Padding(0, 1).            // 内边距
    Height(1).                // 固定高度
    Flex(1)                   // 弹性权重
```

#### Box

容器节点，支持边框和内边距。

```go
func Box(child Node) *boxNode
```

**样式方法**:

```go
rego.Box(child).
    Border(rego.BorderRounded).   // 边框样式
    BorderColor(rego.Cyan).       // 边框颜色
    Padding(1, 2).                // 内边距（垂直, 水平）
    PaddingAll(1, 2, 1, 2).       // 内边距（上, 右, 下, 左）
    Width(40).                    // 固定宽度
    Height(10).                   // 固定高度
    Flex(1).                      // 弹性权重
    Valign(rego.AlignCenter).     // 垂直对齐
    Background(rego.Black)        // 背景色
```

#### Center

居中辅助组件。

```go
func Center(child Node) Node
```

**示例**:

```go
rego.Center(
    rego.Box(
        rego.Text("Centered Content"),
    ).Border(rego.BorderSingle),
)
```

---

### 控制流节点

#### When

条件渲染。

```go
func When(condition bool, node Node) *whenNode
```

**示例**:

```go
rego.When(loading, rego.Spinner(c.Child("spin"), "Loading..."))
rego.When(error != nil, rego.Text(error.Error()).Color(rego.Red))
```

#### WhenElse

条件渲染（带 else 分支）。

```go
func WhenElse(condition bool, trueNode, falseNode Node) *whenElseNode
```

**示例**:

```go
rego.WhenElse(loggedIn,
    UserPanel(c.Child("user")),
    LoginForm(c.Child("login")),
)
```

#### For

列表渲染。

```go
func For[T any](items []T, render func(item T, index int) Node) Node
```

**示例**:

```go
rego.For(todos, func(todo Todo, i int) rego.Node {
    return TodoItem(c.Child("todo", i), todo, i == selected)
})
```

---

### 滚动容器

#### ScrollBox

可滚动容器。

```go
func ScrollBox(c C, child Node) *componentNode
```

**特性**:
- 支持鼠标滚轮
- 自动显示滚动条

#### TailBox

自动滚动到底部的容器（适合日志/聊天）。

```go
func TailBox(c C, child Node) *componentNode
```

**特性**:
- 新内容自动滚动到底部
- 手动向上滚动时暂停自动滚动

---

## 内置组件

### Button

按钮组件。

```go
type ButtonProps struct {
    Label   string   // 按钮文字
    OnClick func()   // 点击回调
    Primary bool     // 是否为主要按钮
}

func Button(c C, props ButtonProps) Node
```

### TextInput

文本输入组件。

```go
type TextInputProps struct {
    Value       string         // 当前值
    Placeholder string         // 占位符
    Label       string         // 标签
    Width       int            // 宽度
    Height      int            // 高度（多行时）
    Multiline   bool           // 是否多行
    Password    bool           // 是否密码模式
    OnChanged   func(string)   // 值变化回调
    OnSubmit    func(string)   // 回车提交回调
}

func TextInput(c C, props TextInputProps) Node
```

### Checkbox

复选框组件。

```go
type CheckboxProps struct {
    Label     string       // 标签
    Checked   bool         // 是否选中
    OnChanged func(bool)   // 状态变化回调
}

func Checkbox(c C, props CheckboxProps) Node
```

### Spinner

加载动画组件。

```go
func Spinner(c C, label string) Node
```

### Markdown

Markdown 渲染组件。

```go
func Markdown(content string) *markdownNode

md := rego.Markdown(content).
    Theme("dark")  // 主题: "dark", "light", "notty"
```

---

## 样式系统

### Style 对象

可复用的样式对象。

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

### 对齐方式

```go
type Align int

const (
    AlignLeft   Align = iota  // 左对齐 / 顶部对齐
    AlignCenter               // 居中对齐
    AlignRight                // 右对齐 / 底部对齐
)
```

---

## 颜色与边框

### 颜色常量

```go
const (
    Default Color = iota  // 终端默认色
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

### 边框样式

```go
const (
    BorderNone    BorderStyle = iota  // 无边框
    BorderSingle                       // ┌─┐
    BorderDouble                       // ╔═╗
    BorderRounded                      // ╭─╮
    BorderThick                        // ┏━┓
)
```

---

## 按键常量

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

## 鼠标事件类型

```go
type MouseEventType int

const (
    MouseEventPress      MouseEventType = iota  // 按下
    MouseEventRelease                            // 释放
    MouseEventClick                              // 点击
    MouseEventMove                               // 移动
    MouseEventScrollUp                           // 向上滚动
    MouseEventScrollDown                         // 向下滚动
)

type MouseButton int

const (
    MouseButtonNone   MouseButton = iota  // 无按键
    MouseButtonLeft                        // 左键
    MouseButtonMiddle                      // 中键
    MouseButtonRight                       // 右键
)
```

---

## 测试工具

Rego 提供测试工具包 `rego/testing`：

```go
import regotest "github.com/erweixin/rego/testing"

func TestMyComponent(t *testing.T) {
    // 创建模拟屏幕
    screen := regotest.NewMockScreen(80, 24)
    
    // 创建测试运行时
    rt := regotest.NewTestRuntime(MyComponent, 80, 24)
    
    // 执行渲染
    rt.Render()
    
    // 获取屏幕内容
    content := rt.Screen.GetContentString()
    
    // 模拟按键
    rt.DispatchKey(rego.KeyEnter, 0, 0)
    
    // 断言结果
    if !strings.Contains(content, "expected text") {
        t.Error("Expected text not found")
    }
}
```

---

## API 速查表

### Hooks

| Hook | 签名 | 说明 |
|------|------|------|
| `Use` | `Use[T](c, key, initial) *State[T]` | 状态管理 |
| `UseEffect` | `UseEffect(c, fn, deps...)` | 副作用 |
| `UseKey` | `UseKey(c, handler)` | 键盘事件 |
| `UseMouse` | `UseMouse(c, handler)` | 鼠标事件 |
| `UseFocus` | `UseFocus(c) FocusState` | 焦点管理 |
| `UseMemo` | `UseMemo[T](c, fn, deps...) T` | 缓存计算 |
| `UseRef` | `UseRef[T](c, initial) *Ref[T]` | 引用 |
| `UseContext` | `UseContext[T](c, ctx) T` | 上下文消费 |
| `UseBridge` | `UseBridge[S,Q,A](c, init) *Bridge` | Agent 通信 |

### 节点

| 类别 | API |
|------|-----|
| **基础** | `Text`, `Empty`, `Spacer`, `Divider`, `Cursor` |
| **布局** | `VStack`, `HStack`, `Box`, `Center` |
| **控制** | `When`, `WhenElse`, `For` |
| **滚动** | `ScrollBox`, `TailBox` |
| **组件** | `Button`, `TextInput`, `Checkbox`, `Spinner`, `Markdown` |

### 上下文方法

| 方法 | 说明 |
|------|------|
| `c.Child(key, i...)` | 获取子组件上下文 |
| `c.Refresh()` | 触发重渲染 |
| `c.Quit()` | 退出应用 |
| `c.Rect()` | 获取组件区域 |
| `c.Wrap(node)` | 包装节点 |

---

*本文档基于 Rego v0.1.0-beta*

