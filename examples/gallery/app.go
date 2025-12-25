package main

import (
	rego "github.com/erweixin/rego"
)

// GalleryApp 是顶层组件，管理核心状态
func GalleryApp(c rego.C) rego.Node {
	// 核心状态提升到顶层
	name := rego.Use(c, "name", "Gopher")
	count := rego.Use(c, "count", 0)
	showModal := rego.Use(c, "showModal", false)

	return rego.VStack(
		// 第一层嵌套：头部
		Header(c.Child("header")),

		rego.Text(""),

		// 第一层嵌套：主体（左右分栏）
		rego.HStack(
			// 左侧：侧边栏组件
			Sidebar(c.Child("sidebar"), count, showModal),

			rego.Text("  "),

			// 右侧：表单内容组件
			Content(c.Child("content"), name),
		).Flex(1),

		rego.Text(""),

		// 第一层嵌套：底部
		Footer(c.Child("footer"), name.Val),

		// 展示 rego.Center 的威力：弹窗
		rego.When(showModal.Val,
			rego.Center(
				rego.Box(
					rego.VStack(
						rego.Text("🎉 恭喜！").Apply(HighlightStyle),
						rego.Divider(),
						rego.Text("这是一个使用 rego.Center 实现的居中弹窗。"),
						rego.Text("它会自动在当前可用空间内双向居中。"),
						rego.Text("测试文本灰色").Apply(DimStyle),
						rego.Button(c.Child("close-modal"), rego.ButtonProps{
							Label: "我知道了",
							OnClick: func() {
								showModal.Set(false)
							},
						}),
					).Gap(1),
				).Apply(ModalStyle),
			),
		),
	)
}
