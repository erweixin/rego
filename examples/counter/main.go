package main

import (
	"fmt"
	"log"

	"github.com/erweixin/rego"
)

// =============================================================================
// Counter 示例 - 展示状态管理、按钮组件和多面板布局
// =============================================================================

func App(c rego.C) rego.Node {
	activePanel := rego.Use(c, "activePanel", 0) // 0: 计数器, 1: 历史记录

	rego.UseKey(c, func(key rego.Key, r rune) {
		switch key {
		case rego.KeyTab:
			activePanel.Set((activePanel.Val + 1) % 2)
		}
		switch r {
		case '1':
			activePanel.Set(0)
		case '2':
			activePanel.Set(1)
		case 'q':
			c.Quit()
		}
	})

	return rego.VStack(
		// 顶部标题栏
		Header(c.Child("header")),

		rego.Text(""),

		// 主体区域
		rego.HStack(
			// 左侧：计数器面板
			CounterPanel(c.Child("counter"), activePanel.Val == 0),

			rego.Text("  "),

			// 右侧：历史记录面板
			HistoryPanel(c.Child("history"), activePanel.Val == 1),
		).Flex(1),

		rego.Text(""),

		// 底部状态栏
		Footer(c.Child("footer")),
	).Padding(1, 2)
}

// =============================================================================
// Header 组件
// =============================================================================

func Header(c rego.C) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("🎯 Rego Counter").Bold().Color(rego.Cyan),
			rego.Spacer(),
			rego.Text("[Tab] 切换面板").Dim(),
			rego.Text("  "),
			rego.Text("[q] 退出").Dim(),
		),
	).Border(rego.BorderDouble).BorderColor(rego.Cyan).Padding(0, 1)
}

// =============================================================================
// CounterPanel 组件 - 计数器主面板
// =============================================================================

func CounterPanel(c rego.C, active bool) rego.Node {
	count := rego.Use(c, "count", 0)
	step := rego.Use(c, "step", 1)

	// 只在激活时处理面板特定的按键
	if active {
		rego.UseKey(c, func(key rego.Key, r rune) {
			switch r {
			case '+', '=':
				count.Set(count.Val + step.Val)
			case '-', '_':
				count.Set(count.Val - step.Val)
			case 'r':
				count.Set(0)
			}
			switch key {
			case rego.KeyUp:
				step.Set(step.Val + 1)
			case rego.KeyDown:
				if step.Val > 1 {
					step.Set(step.Val - 1)
				}
			}
		})
	}

	borderColor := rego.Gray
	if active {
		borderColor = rego.Green
	}

	// 计数值的颜色
	countColor := rego.White
	if count.Val > 0 {
		countColor = rego.Green
	} else if count.Val < 0 {
		countColor = rego.Red
	}

	return rego.Box(
		rego.VStack(
			rego.HStack(
				rego.Text("📊 计数器").Bold(),
				rego.When(active, rego.Text(" ●").Color(rego.Green)),
			),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			// 大号计数显示
			rego.Box(
				rego.Text(fmt.Sprintf(" %d ", count.Val)).Bold().Color(countColor),
			).Border(rego.BorderRounded).BorderColor(countColor).Padding(1, 4),

			rego.Text(""),

			// 步进值显示
			rego.HStack(
				rego.Text("步进值: "),
				rego.Text(fmt.Sprintf("%d", step.Val)).Bold().Color(rego.Yellow),
				rego.Text(" (↑/↓ 调整)").Dim(),
			),

			rego.Text(""),

			// 操作按钮
			rego.HStack(
				rego.Button(c.Child("btn-add"), rego.ButtonProps{
					Label:   " + 增加 ",
					Primary: true,
					OnClick: func() { count.Set(count.Val + step.Val) },
				}),
				rego.Text(" "),
				rego.Button(c.Child("btn-sub"), rego.ButtonProps{
					Label:   " - 减少 ",
					OnClick: func() { count.Set(count.Val - step.Val) },
				}),
				rego.Text(" "),
				rego.Button(c.Child("btn-reset"), rego.ButtonProps{
					Label:   " ↺ 重置 ",
					OnClick: func() { count.Set(0) },
				}),
			),

			rego.Spacer(),

			// 快捷键提示
			rego.Text("─────────────────────────────").Dim(),
			rego.Text("[+/-] 增减  [r] 重置").Dim(),
		),
	).Flex(1).Border(rego.BorderSingle).BorderColor(borderColor).Padding(1, 2)
}

// =============================================================================
// HistoryPanel 组件 - 历史记录面板
// =============================================================================

func HistoryPanel(c rego.C, active bool) rego.Node {
	history := rego.Use(c, "history", []int{0})
	selected := rego.Use(c, "selected", 0)

	// 监听计数器变化（通过共享 context）
	// 这里简化处理，只展示布局

	if active {
		rego.UseKey(c, func(key rego.Key, r rune) {
			switch key {
			case rego.KeyUp:
				if selected.Val > 0 {
					selected.Set(selected.Val - 1)
				}
			case rego.KeyDown:
				if selected.Val < len(history.Val)-1 {
					selected.Set(selected.Val + 1)
				}
			}
			switch r {
			case 'c':
				history.Set([]int{0})
				selected.Set(0)
			case 'a':
				// 添加一个随机值到历史
				newVal := (len(history.Val) + 1) * 10
				history.Set(append(history.Val, newVal))
			}
		})
	}

	borderColor := rego.Gray
	if active {
		borderColor = rego.Green
	}

	return rego.Box(
		rego.VStack(
			rego.HStack(
				rego.Text("📜 历史记录").Bold(),
				rego.When(active, rego.Text(" ●").Color(rego.Green)),
			),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			// 历史列表
			rego.ScrollBox(c.Child("scroll"),
				rego.For(history.Val, func(val int, i int) rego.Node {
					prefix := "  "
					color := rego.White
					if i == selected.Val && active {
						prefix = "▸ "
						color = rego.Green
					}
					return rego.Text(fmt.Sprintf("%s#%d: %d", prefix, i+1, val)).Color(color)
				}),
			).Flex(1),

			rego.Spacer(),

			// 操作按钮
			rego.HStack(
				rego.Button(c.Child("btn-add-history"), rego.ButtonProps{
					Label: " + 添加 ",
					OnClick: func() {
						newVal := (len(history.Val) + 1) * 10
						history.Set(append(history.Val, newVal))
					},
				}),
				rego.Text(" "),
				rego.Button(c.Child("btn-clear"), rego.ButtonProps{
					Label: " ✕ 清空 ",
					OnClick: func() {
						history.Set([]int{0})
						selected.Set(0)
					},
				}),
			),

			rego.Text(""),
			rego.Text("─────────────────────────────").Dim(),
			rego.Text("[↑/↓] 选择  [a] 添加  [c] 清空").Dim(),
		),
	).Flex(1).Border(rego.BorderSingle).BorderColor(borderColor).Padding(1, 2)
}

// =============================================================================
// Footer 组件
// =============================================================================

func Footer(c rego.C) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("状态: ").Dim(),
			rego.Text("就绪").Color(rego.Green),
			rego.Spacer(),
			rego.Text("[1] 计数器  [2] 历史记录").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(rego.Gray).Padding(0, 1)
}

func main() {
	if err := rego.Run(App); err != nil {
		log.Fatal(err)
	}
}
