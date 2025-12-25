package main

import (
	"fmt"
	rego "github.com/erweixin/rego"
)

func App(c rego.C) rego.Node {
	username := rego.Use(c, "username", "")
	password := rego.Use(c, "password", "")
	bio := rego.Use(c, "bio", "这是一段很长很长很长很长很长很长很长很长的自我介绍，用来测试滚动。")
	submitted := rego.Use(c, "submitted", false)

	rego.UseKey(c, func(key rego.Key, r rune) {
		if r == 'q' {
			c.Quit()
		}
	})

	return rego.Box(
		rego.VStack(
			rego.Text("📝 用户注册表单").Bold().Color(rego.Cyan),
			rego.Text("使用 Tab 切换焦点，鼠标滚轮滚动下方区域").Dim(),
			rego.Text(""),

			// 输入框演示
			rego.TextInput(c.Child("input-user"), rego.TextInputProps{
				Label:       "用户名",
				Placeholder: "请输入用户名...",
				Value:       username.Val,
				Width:       40,
				OnChanged:   func(s string) { username.Set(s) },
			}),

			rego.Text(""),

			rego.TextInput(c.Child("input-pwd"), rego.TextInputProps{
				Label:       "密码",
				Placeholder: "请输入密码...",
				Value:       password.Val,
				Width:       40,
				Password:    true,
				OnChanged:   func(s string) { password.Set(s) },
			}),

			rego.Text(""),

			rego.TextInput(c.Child("input-bio"), rego.TextInputProps{
				Label:       "个人简介 (多行输入)",
				Placeholder: "介绍一下你自己...",
				Value:       bio.Val,
				Width:       50,
				Height:      6,
				Multiline:   true,
				OnChanged:   func(s string) { bio.Set(s) },
			}),

			rego.Text(""),

			// 滚动区域演示
			rego.Text("📜 滚动协议区域:").Bold(),
			rego.Box(
				rego.ScrollBox(c.Child("scroller"),
					rego.VStack(
						rego.Text("1. 请确保你已经阅读本协议。"),
						rego.Text("2. Rego 是一个好用的框架。"),
						rego.Text("3. 你可以自由地使用它。"),
						rego.Text("4. 鼠标滚动可以查看更多内容。"),
						rego.Text("5. 这里是填充行 A..."),
						rego.Text("6. 这里是填充行 B..."),
						rego.Text("7. 这里是填充行 C..."),
						rego.Text("8. 这里是填充行 D..."),
						rego.Text("9. 这里是填充行 E..."),
						rego.Text("10. 自我介绍: "+bio.Val),
						rego.Text("11. 更多行 1..."),
						rego.Text("12. 更多行 2..."),
						rego.Text("13. 更多行 3..."),
						rego.Text("14. 更多行 4..."),
						rego.Text("15. 协议结束。"),
					),
				),
			).Height(6).Border(rego.BorderSingle).BorderColor(rego.Gray),

			rego.Text(""),

			Button(c.Child("submit"), "提交表单", func() {
				submitted.Set(true)
			}),

			rego.When(submitted.Val,
				rego.Text(fmt.Sprintf("\n✅ 提交成功！欢迎，%s", username.Val)).Color(rego.Green),
			),

			rego.Spacer(),
			rego.Text("按 [q] 退出").Dim(),
		),
	).Padding(1, 2).Width(60).Height(28).Border(rego.BorderSingle)
}

// 复用之前的 Button 组件逻辑，简单实现一个
func Button(c rego.C, label string, onClick func()) rego.Node {
	focus := rego.UseFocus(c)
	rego.UseMouse(c, func(ev rego.MouseEvent) {
		if ev.Type == rego.MouseEventClick && c.Rect().Contains(ev.X, ev.Y) {
			onClick()
			focus.Focus()
		}
	})

	return c.Wrap(rego.Box(
		rego.Text(label).Color(rego.If(focus.IsFocused, rego.Black, rego.White)).
			Background(rego.If(focus.IsFocused, rego.Cyan, rego.Default)),
	).Border(rego.BorderSingle).Padding(0, 1))
}

func main() {
	if err := rego.Run(App); err != nil {
		panic(err)
	}
}
