package main

import (
	"fmt"
	"time"

	rego "github.com/erweixin/rego"
)

func Sidebar(c rego.C, count *rego.State[int], showModal *rego.State[bool]) rego.Node {
	return rego.Box(
		rego.VStack(
			rego.Text("控制中心").Apply(SidebarTitleStyle),
			rego.Text(""),

			// 第二层嵌套：计数器部分
			CounterSection(c.Child("counter"), count),

			rego.Text(""),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			// 第二层嵌套：操作部分
			ActionSection(c.Child("actions")),

			rego.Text(""),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			// 展示 rego.Center 触发器
			rego.Button(c.Child("btn-modal"), rego.ButtonProps{
				Label:   "打开居中弹窗",
				OnClick: func() { showModal.Set(true) },
			}),

			rego.Spacer(),
			rego.Text("Tab 切换焦点").Apply(DimStyle),
		),
	).Apply(CardStyle).Flex(1)
}

func CounterSection(c rego.C, count *rego.State[int]) rego.Node {
	return rego.VStack(
		rego.HStack(
			rego.Text("计数: "),
			rego.Button(c.Child("btn-add"), rego.ButtonProps{
				Label:   "增加",
				OnClick: func() { count.Set(count.Val + 1) },
			}),
			rego.Text(" "),
			rego.Button(c.Child("btn-reset"), rego.ButtonProps{
				Label:   "清零",
				Primary: true,
				OnClick: func() { count.Set(0) },
			}),
		),
		rego.Text(fmt.Sprintf("当前数值: %d", count.Val)).Apply(DimStyle),
	)
}

func ActionSection(c rego.C) rego.Node {
	showSpinner := rego.Use(c, "showSpinner", false)
	checkboxChecked := rego.Use(c, "checkboxChecked", false)

	// 第三层嵌套：处理副作用逻辑也可以封装在子组件内
	rego.UseEffect(c, func() func() {
		if showSpinner.Val {
			timer := time.AfterFunc(2*time.Second, func() {
				showSpinner.Set(false)
				c.Refresh()
			})
			return func() { timer.Stop() }
		}
		return nil
	}, showSpinner.Val)

	return rego.VStack(
		rego.Checkbox(c.Child("check"), rego.CheckboxProps{
			Label:   "启用自动同步",
			Checked: checkboxChecked.Val,
			OnChanged: func(checked bool) {
				checkboxChecked.Set(checked)
			},
		}),
		rego.HStack(
			rego.Button(c.Child("btn-load"), rego.ButtonProps{
				Label:   "执行任务",
				OnClick: func() { showSpinner.Set(true) },
			}),
			rego.When(showSpinner.Val,
				rego.HStack(rego.Text("  "), rego.Spinner(c.Child("spin"), "加载中")),
			),
		),
	).Gap(10) // 使用 Gap 代替手动插入 Text("")
}
