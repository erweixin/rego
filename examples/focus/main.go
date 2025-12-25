package main

import (
	"unicode"

	rego "github.com/erweixin/rego"
)

// =============================================================================
// App - 展示 UseFocus 焦点系统
// =============================================================================

func App(c rego.C) rego.Node {
	return rego.VStack(
		rego.Text(""),
		rego.Text("  🎯 Rego Focus System Demo").Bold().Color(rego.Cyan),
		rego.Text("  使用 Tab/Shift+Tab 在输入框之间切换焦点").Dim(),
		rego.Text(""),

		// 三个可聚焦的输入框
		InputField(c.Child("name"), "姓名", "请输入您的姓名"),
		rego.Text(""),
		InputField(c.Child("email"), "邮箱", "请输入您的邮箱"),
		rego.Text(""),
		InputField(c.Child("message"), "留言", "请输入您的留言"),

		rego.Spacer(),

		// 底部说明
		rego.Text("  ─────────────────────────────────────────"),
		rego.Text("  [Tab] 下一个  [Shift+Tab] 上一个  [Ctrl+C] 退出").Dim(),
	)
}

// =============================================================================
// InputField - 可聚焦的输入框组件
// =============================================================================

func InputField(c rego.C, label, placeholder string) rego.Node {
	value := rego.Use(c, "value", "")
	focus := rego.UseFocus(c) // 声明可聚焦

	// 只在聚焦时处理键盘输入
	rego.UseKey(c, func(key rego.Key, r rune) {
		if !focus.IsFocused {
			return // 未聚焦，不处理
		}

		switch {
		case key == rego.KeyBackspace:
			if len(value.Val) > 0 {
				// 删除最后一个字符（支持 UTF-8）
				runes := []rune(value.Val)
				value.Set(string(runes[:len(runes)-1]))
			}
		case key == rego.KeyEsc:
			value.Set("") // 清空
		case unicode.IsPrint(r):
			value.Set(value.Val + string(r))
		}
	})

	// 根据焦点状态设置样式
	borderColor := rego.Gray
	labelColor := rego.Gray
	if focus.IsFocused {
		borderColor = rego.Green
		labelColor = rego.Green
	}

	// 显示内容
	displayText := value.Val
	if displayText == "" && !focus.IsFocused {
		displayText = placeholder
	}

	// 光标
	cursor := ""
	if focus.IsFocused {
		cursor = "▌"
	}

	return rego.Box(
		rego.VStack(
			rego.HStack(
				rego.Text(label).Bold().Color(labelColor),
				rego.When(focus.IsFocused,
					rego.Text(" ✓").Color(rego.Green),
				),
			),
			rego.HStack(
				rego.WhenElse(value.Val == "" && !focus.IsFocused,
					rego.Text(displayText).Dim(),
					rego.Text(displayText).Color(rego.White),
				),
				rego.Text(cursor).Color(rego.Green),
			),
		),
	).Width(50).Border(rego.BorderSingle).BorderColor(borderColor).Padding(0, 1)
}

func main() {
	if err := rego.Run(App); err != nil {
		panic(err)
	}
}
