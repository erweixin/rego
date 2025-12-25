package main

import (
	"log"

	"github.com/erweixin/rego"
)

// =============================================================================
// Hello World 示例 - 展示 Rego 的基础布局和样式系统
// =============================================================================

func App(c rego.C) rego.Node {
	// 键盘事件处理
	rego.UseKey(c, func(key rego.Key, r rune) {
		if r == 'q' || key == rego.KeyCtrlC {
			c.Quit()
		}
	})

	return rego.VStack(
		// 顶部标题栏
		rego.Box(
			rego.HStack(
				rego.Text("🎯 Hello, Rego!").Bold().Color(rego.Cyan),
				rego.Spacer(),
				rego.Text("v0.1.0").Dim(),
			),
		).Border(rego.BorderDouble).BorderColor(rego.Cyan).Padding(0, 1),

		rego.Text(""),

		// 主要内容区域
		rego.HStack(
			// 左侧介绍卡片
			rego.Box(
				rego.VStack(
					rego.Text("📦 框架特点").Bold().Color(rego.Yellow),
					rego.Divider().Color(rego.Gray),
					rego.Text(""),
					rego.Text("• React Hooks 风格").Color(rego.White),
					rego.Text("• 声明式 UI").Color(rego.White),
					rego.Text("• 类型安全").Color(rego.White),
					rego.Text("• 组件化开发").Color(rego.White),
					rego.Text("• 灵活的布局系统").Color(rego.White),
					rego.Spacer(),
				),
			).Border(rego.BorderRounded).BorderColor(rego.Yellow).Padding(1, 2).Flex(1),

			rego.Text("  "),

			// 右侧代码示例卡片
			rego.Box(
				rego.VStack(
					rego.Text("💻 快速上手").Bold().Color(rego.Green),
					rego.Divider().Color(rego.Gray),
					rego.Text(""),
					rego.Text("func App(c rego.C) rego.Node {").Color(rego.Cyan),
					rego.Text("    return rego.Text(\"Hello!\")").Color(rego.White),
					rego.Text("}").Color(rego.Cyan),
					rego.Text(""),
					rego.Text("func main() {").Color(rego.Cyan),
					rego.Text("    rego.Run(App)").Color(rego.White),
					rego.Text("}").Color(rego.Cyan),
					rego.Spacer(),
				),
			).Border(rego.BorderRounded).BorderColor(rego.Green).Padding(1, 2).Flex(1),
		),

		rego.Text(""),

		// 底部操作栏
		rego.Box(
			rego.HStack(
				rego.Text("欢迎使用 Rego TUI 框架！").Color(rego.White),
				rego.Spacer(),
				rego.Text("[q] 退出").Dim(),
			),
		).Border(rego.BorderSingle).BorderColor(rego.Gray).Padding(0, 1),
	).Padding(1, 2)
}

func main() {
	if err := rego.Run(App); err != nil {
		log.Fatal(err)
	}
}
