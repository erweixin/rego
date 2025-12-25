package main

import (
	rego "github.com/erweixin/rego"
)

func Content(c rego.C, name *rego.State[string]) rego.Node {
	return rego.Box(
		rego.VStack(
			rego.Text("设置面板").Bold().Underline().Color(rego.Yellow),
			rego.Text(""),

			// 第二层嵌套：用户信息表单
			UserForm(c.Child("user-form"), name),
		),
	).
		Border(rego.BorderSingle).
		Padding(1, 2).
		Flex(1).
		Valign(rego.AlignCenter) // 展示垂直居中对齐
}

func UserForm(c rego.C, name *rego.State[string]) rego.Node {
	password := rego.Use(c, "password", "")
	bio := rego.Use(c, "bio", "这个简介是在 UserForm 子组件内部管理的。")

	return rego.VStack(
		// 第三层嵌套：基础信息
		rego.TextInput(c.Child("input-name"), rego.TextInputProps{
			Label:       "用户名",
			Value:       name.Val,
			Placeholder: "请输入",
			OnChanged:   func(v string) { name.Set(v) },
		}),

		rego.Text(""),
		rego.Divider().Char('.').Color(rego.Gray),
		rego.Text(""),

		// 第三层嵌套：敏感信息
		rego.TextInput(c.Child("input-pwd"), rego.TextInputProps{
			Label:       "密码",
			Value:       password.Val,
			Password:    true,
			Placeholder: "请输入",
			OnChanged:   func(v string) { password.Set(v) },
		}),

		rego.Text(""),
		rego.Divider().Char('.').Color(rego.Gray),
		rego.Text(""),

		// 第三层嵌套：详细信息
		rego.TextInput(c.Child("input-bio"), rego.TextInputProps{
			Label:     "个人简介",
			Value:     bio.Val,
			Multiline: true,
			Height:    5,
			OnChanged: func(v string) { bio.Set(v) },
		}),
	)
}
