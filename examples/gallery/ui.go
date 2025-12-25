package main

import (
	"fmt"

	rego "github.com/erweixin/rego"
)

func Header(c rego.C) rego.Node {
	return rego.Box(
		rego.Text("🎨 REGO NESTED GALLERY").Apply(TitleStyle),
	).Border(rego.BorderDouble).BorderColor(rego.Cyan).Padding(0, 1)
}

func Footer(c rego.C, userName string) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("按 Ctrl+C 退出").Apply(DimStyle),
			rego.Spacer(),
			rego.Text(fmt.Sprintf("当前用户: %s", userName)).Apply(HighlightStyle),
		),
	).Apply(CardStyle).Padding(0, 1) // 复用 CardStyle 但覆盖 Padding
}
