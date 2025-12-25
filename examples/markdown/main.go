package main

import (
	rego "github.com/erweixin/rego"
)

// =============================================================================
// Markdown 示例 - 展示 Markdown 渲染、主题切换、Tab 切换预览模式
// =============================================================================

// 主题配置
type Theme struct {
	Name       string
	Style      string // glamour 主题
	Primary    rego.Color
	Background rego.Color
	Border     rego.Color
}

var themes = []Theme{
	{Name: "🌙 暗色", Style: "dark", Primary: rego.Cyan, Background: rego.Default, Border: rego.Cyan},
	{Name: "☀️ 亮色", Style: "light", Primary: rego.Blue, Background: rego.Default, Border: rego.Blue},
	{Name: "🔮 无样式", Style: "notty", Primary: rego.Magenta, Background: rego.Default, Border: rego.Magenta},
}

// 示例内容
var sampleContents = []struct {
	title   string
	content string
}{
	{
		title: "📖 基础语法",
		content: `# Markdown 基础语法

这是一个 **Markdown** 渲染示例，由 [glamour](https://github.com/charmbracelet/glamour) 提供支持。

## 文本格式

- **加粗文本**
- *斜体文本*
- ~~删除线~~
- ` + "`行内代码`" + `

## 链接

访问 [Rego 仓库](https://github.com/erweixin/rego) 了解更多。

> 这是一个引用块。
> 可以包含多行内容。

---
以上就是基本的 Markdown 语法。
`,
	},
	{
		title: "💻 代码高亮",
		content: `# 代码高亮示例

Markdown 支持多种编程语言的语法高亮。

## Go 代码

` + "```go" + `
package main

import (
    "fmt"
    "github.com/erweixin/rego"
)

func App(c rego.C) rego.Node {
    count := rego.Use(c, "count", 0)
    
    return rego.VStack(
        rego.Text(fmt.Sprintf("Count: %d", count.Val)),
        rego.Button(c.Child("btn"), rego.ButtonProps{
            Label: "增加",
            OnClick: func() { count.Set(count.Val + 1) },
        }),
    )
}

func main() {
    rego.Run(App)
}
` + "```" + `

## JavaScript 代码

` + "```javascript" + `
const greeting = (name) => {
    console.log(` + "`Hello, ${name}!`" + `);
};

greeting("World");
` + "```" + `

## Shell 命令

` + "```bash" + `
# 安装 Rego
go get github.com/erweixin/rego

# 运行示例
go run examples/markdown/main.go
` + "```" + `
`,
	},
	{
		title: "📋 列表与表格",
		content: `# 列表与表格

## 有序列表

1. 第一步：安装 Go
2. 第二步：创建项目
3. 第三步：引入 Rego
4. 第四步：编写组件
5. 第五步：运行应用

## 无序列表

- 项目结构
  - main.go
  - components/
    - header.go
    - sidebar.go
  - styles/
    - theme.go

## 任务列表

- [x] 完成项目初始化
- [x] 添加基础组件
- [ ] 实现主题切换
- [ ] 添加更多示例

## 表格

| 功能 | 描述 | 状态 |
|------|------|------|
| UseState | 状态管理 | ✅ 完成 |
| UseEffect | 副作用处理 | ✅ 完成 |
| UseMemo | 计算缓存 | ✅ 完成 |
| UseContext | 上下文共享 | ✅ 完成 |

`,
	},
}

func App(c rego.C) rego.Node {
	themeIndex := rego.Use(c, "theme", 0)
	contentIndex := rego.Use(c, "content", 0)
	showSidebar := rego.Use(c, "sidebar", true)

	currentTheme := themes[themeIndex.Val]
	currentContent := sampleContents[contentIndex.Val]

	// 键盘事件
	rego.UseKey(c, func(key rego.Key, r rune) {
		switch key {
		case rego.KeyLeft:
			if contentIndex.Val > 0 {
				contentIndex.Set(contentIndex.Val - 1)
			}
		case rego.KeyRight:
			if contentIndex.Val < len(sampleContents)-1 {
				contentIndex.Set(contentIndex.Val + 1)
			}
		}
		switch r {
		// 使用 a/d 或 h/l 切换内容页面
		case 'a', 'h':
			if contentIndex.Val > 0 {
				contentIndex.Set(contentIndex.Val - 1)
			}
		case 'd', 'l':
			if contentIndex.Val < len(sampleContents)-1 {
				contentIndex.Set(contentIndex.Val + 1)
			}
		// 使用 w/e/r 切换主题
		case 'w':
			themeIndex.Set(0)
		case 'e':
			themeIndex.Set(1)
		case 'r':
			themeIndex.Set(2)
		case 's':
			showSidebar.Set(!showSidebar.Val)
		case 'q':
			c.Quit()
		}
	})

	return rego.VStack(
		// 顶部标题栏
		Header(c.Child("header"), currentTheme),

		rego.Text(""),

		// Tab 栏
		TabBar(c.Child("tabs"), contentIndex.Val, sampleContents, currentTheme),

		rego.Text(""),

		// 主体内容
		rego.HStack(
			// 左侧：Markdown 预览
			MarkdownPreview(c.Child("preview"), currentContent.content, currentTheme),

			// 右侧：侧边栏（可折叠）
			rego.When(showSidebar.Val,
				rego.HStack(
					rego.Text("  "),
					Sidebar(c.Child("sidebar"), themeIndex.Val, themeIndex.Set, currentTheme),
				),
			),
		).Flex(1),

		rego.Text(""),

		// 底部状态栏
		Footer(c.Child("footer"), currentTheme, showSidebar.Val),
	).Padding(1, 2)
}

// =============================================================================
// Header 组件
// =============================================================================

func Header(c rego.C, theme Theme) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("📄 Rego Markdown Viewer").Bold().Color(theme.Primary),
			rego.Spacer(),
			rego.Text(theme.Name).Color(theme.Primary),
		),
	).Border(rego.BorderDouble).BorderColor(theme.Border).Padding(0, 1)
}

// =============================================================================
// TabBar 组件
// =============================================================================

func TabBar(c rego.C, activeIndex int, contents []struct {
	title   string
	content string
}, theme Theme) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.For(contents, func(item struct {
				title   string
				content string
			}, i int) rego.Node {
				isActive := i == activeIndex
				text := rego.Text(" " + item.title + " ")

				if isActive {
					return text.Bold().Color(rego.Black).Background(theme.Primary)
				}
				return text.Color(rego.Gray)
			}),
			rego.Spacer(),
			rego.Text("[←/→] 或 [a/d] 切换").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(rego.Gray).Padding(0, 1)
}

// =============================================================================
// MarkdownPreview 组件
// =============================================================================

func MarkdownPreview(c rego.C, content string, theme Theme) rego.Node {
	return rego.Box(
		rego.ScrollBox(c.Child("scroll"),
			rego.Markdown(content).Theme(theme.Style),
		).Flex(1),
	).Border(rego.BorderSingle).BorderColor(theme.Border).Padding(1, 2).Flex(3)
}

// =============================================================================
// Sidebar 组件
// =============================================================================

func Sidebar(c rego.C, themeIndex int, setTheme func(int), currentTheme Theme) rego.Node {
	return rego.Box(
		rego.VStack(
			rego.Text("🎨 主题设置").Bold().Color(currentTheme.Primary),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			// 主题选择
			rego.For(themes, func(t Theme, i int) rego.Node {
				isActive := i == themeIndex
				prefix := "  "
				if isActive {
					prefix = "▸ "
				}

				text := rego.Text(prefix + t.Name)
				if isActive {
					return text.Bold().Color(t.Primary)
				}
				return text.Color(rego.Gray)
			}),

			rego.Text(""),
			rego.Text("按 [w/e/r] 切换主题").Dim(),

			rego.Spacer(),

			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			// 帮助信息
			rego.Text("⌨️ 快捷键").Bold().Color(currentTheme.Primary),
			rego.Text(""),
			rego.Text("[←/→] 切换内容").Dim(),
			rego.Text("[a/d] 上/下一页").Dim(),
			rego.Text("[w/e/r] 切换主题").Dim(),
			rego.Text("[s] 显示/隐藏侧栏").Dim(),
			rego.Text("[q] 退出").Dim(),
		),
	).Width(25).Border(rego.BorderSingle).BorderColor(currentTheme.Border).Padding(1, 1)
}

// =============================================================================
// Footer 组件
// =============================================================================

func Footer(c rego.C, theme Theme, sidebarVisible bool) rego.Node {
	sidebarStatus := "开"
	if !sidebarVisible {
		sidebarStatus = "关"
	}

	return rego.Box(
		rego.HStack(
			rego.Text("侧栏: ").Dim(),
			rego.Text(sidebarStatus).Color(rego.If(sidebarVisible, rego.Green, rego.Gray)),
			rego.Text("  "),
			rego.Text("[s] 切换侧栏").Dim(),
			rego.Spacer(),
			rego.Text("Powered by glamour").Dim(),
			rego.Text("  "),
			rego.Text("[q] 退出").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(rego.Gray).Padding(0, 1)
}

func main() {
	if err := rego.Run(App); err != nil {
		panic(err)
	}
}
