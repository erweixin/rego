package main

import rego "github.com/erweixin/rego"

// 定义全局可复用的样式
var (
	// 标题样式
	TitleStyle = rego.NewStyle().
			Bold().
			Foreground(rego.Cyan).
			Align(rego.AlignCenter)

	// 边栏标题样式
	SidebarTitleStyle = rego.NewStyle().
				Bold().
				Underline().
				Foreground(rego.Yellow)

	// 卡片/容器样式
	CardStyle = rego.NewStyle().
			Border(rego.BorderSingle).
			Padding(1, 2)

	// 强调文本样式
	HighlightStyle = rego.NewStyle().
			Bold().
			Foreground(rego.Green)

	// 次要/辅助文本样式
	DimStyle = rego.NewStyle().
			Dim().
			Italic()

	// 弹窗样式
	ModalStyle = rego.NewStyle().
			Border(rego.BorderRounded).
			BorderColor(rego.Green).
			Padding(1, 2).
			Width(50)
)
