package main

import (
	rego "github.com/erweixin/rego"
)

// =============================================================================
// Theme Context - 主题上下文
// =============================================================================

// Theme 主题配置
type Theme struct {
	Name       string
	Primary    rego.Color
	Secondary  rego.Color
	Background rego.Color
	Text       rego.Color
	Border     rego.Color
}

// 预定义主题
var (
	DarkTheme = Theme{
		Name:       "🌙 暗色主题",
		Primary:    rego.Cyan,
		Secondary:  rego.Blue,
		Background: rego.Black,
		Text:       rego.White,
		Border:     rego.Gray,
	}

	LightTheme = Theme{
		Name:       "☀️ 亮色主题",
		Primary:    rego.Blue,
		Secondary:  rego.Cyan,
		Background: rego.White,
		Text:       rego.Black,
		Border:     rego.Gray,
	}

	NeonTheme = Theme{
		Name:       "🌈 霓虹主题",
		Primary:    rego.Magenta,
		Secondary:  rego.Green,
		Background: rego.Black,
		Text:       rego.Yellow,
		Border:     rego.Magenta,
	}
)

// 创建主题 Context
var ThemeContext = rego.CreateContext(DarkTheme)

// =============================================================================
// App - 主应用
// =============================================================================

func App(c rego.C) rego.Node {
	themeIndex := rego.Use(c, "themeIndex", 0)
	themes := []Theme{DarkTheme, LightTheme, NeonTheme}
	currentTheme := themes[themeIndex.Val]

	// 键盘切换主题
	rego.UseKey(c, func(key rego.Key, r rune) {
		switch {
		case r == '1':
			themeIndex.Set(0)
		case r == '2':
			themeIndex.Set(1)
		case r == '3':
			themeIndex.Set(2)
		case key == rego.KeyTab:
			themeIndex.Update(func(i int) int { return (i + 1) % 3 })
		case r == 'q':
			c.Quit()
		}
	})

	// 使用 ThemeContext.Provide 提供主题给所有子组件
	return ThemeContext.Provide(c, currentTheme,
		rego.VStack(
			Header(c.Child("header")),
			rego.Text(""),
			Content(c.Child("content")),
			rego.Text(""),
			ThemeSwitcher(c.Child("switcher"), themeIndex.Val),
			rego.Spacer(),
			Footer(c.Child("footer")),
		),
	)
}

// =============================================================================
// Header - 使用主题的头部组件
// =============================================================================

func Header(c rego.C) rego.Node {
	theme := rego.UseContext(c, ThemeContext) // 从 Context 获取主题

	return rego.Box(
		rego.HStack(
			rego.Text("🎨 Rego Theme Demo").Bold().Color(theme.Primary),
			rego.Spacer(),
			rego.Text(theme.Name).Color(theme.Secondary),
		),
	).Border(rego.BorderSingle).BorderColor(theme.Border).Padding(0, 1)
}

// =============================================================================
// Content - 内容区域
// =============================================================================

func Content(c rego.C) rego.Node {
	theme := rego.UseContext(c, ThemeContext)

	return rego.VStack(
		rego.Text("  这是一个演示 Context API 的示例").Color(theme.Text),
		rego.Text(""),
		rego.Text("  主题颜色会自动传递给所有子组件：").Color(theme.Text),
		rego.Text(""),
		Card(c.Child("card1"), "📦 组件 A", "深层嵌套的组件也能获取主题"),
		rego.Text(""),
		Card(c.Child("card2"), "🔧 组件 B", "无需手动传递 props"),
		rego.Text(""),
		Card(c.Child("card3"), "✨ 组件 C", "Context 让状态共享变得简单"),
	)
}

// =============================================================================
// Card - 卡片组件（深层嵌套，自动获取主题）
// =============================================================================

func Card(c rego.C, title, description string) rego.Node {
	theme := rego.UseContext(c, ThemeContext) // 深层组件也能获取主题！

	return rego.Box(
		rego.VStack(
			rego.Text(title).Bold().Color(theme.Primary),
			rego.Text(description).Color(theme.Text),
		),
	).Width(50).Border(rego.BorderSingle).BorderColor(theme.Border).Padding(0, 1)
}

// =============================================================================
// ThemeSwitcher - 主题切换器
// =============================================================================

func ThemeSwitcher(c rego.C, currentIndex int) rego.Node {
	theme := rego.UseContext(c, ThemeContext)

	themes := []string{"[1] 暗色", "[2] 亮色", "[3] 霓虹"}

	return rego.Box(
		rego.HStack(
			rego.Text("  切换主题: ").Color(theme.Text),
			rego.For(themes, func(name string, i int) rego.Node {
				if i == currentIndex {
					return rego.Text(name + " ").Bold().Color(theme.Primary)
				}
				return rego.Text(name + " ").Dim()
			}),
		),
	).Border(rego.BorderSingle).BorderColor(theme.Border).Padding(0, 1)
}

// =============================================================================
// Footer - 底部
// =============================================================================

func Footer(c rego.C) rego.Node {
	theme := rego.UseContext(c, ThemeContext)

	return rego.Box(
		rego.HStack(
			rego.Text("[1/2/3] 切换主题").Dim(),
			rego.Text("  "),
			rego.Text("[Tab] 循环").Dim(),
			rego.Spacer(),
			rego.Text("[q] 退出").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(theme.Border).Padding(0, 1)
}

func main() {
	if err := rego.Run(App); err != nil {
		panic(err)
	}
}
