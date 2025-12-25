package main

import (
	"fmt"
	"log"

	"github.com/erweixin/rego"
)

// =============================================================================
// Todo 示例 - 展示焦点管理、Tab 切换、组件化开发
// =============================================================================

// Todo 任务结构
type Todo struct {
	Text      string
	Completed bool
}

// 过滤类型
const (
	FilterAll       = 0
	FilterActive    = 1
	FilterCompleted = 2
)

func App(c rego.C) rego.Node {
	todos := rego.Use(c, "todos", []Todo{
		{Text: "学习 Go 语言", Completed: true},
		{Text: "写 Rego 应用", Completed: false},
		{Text: "喝杯咖啡", Completed: false},
	})
	filter := rego.Use(c, "filter", FilterAll)
	activePanel := rego.Use(c, "activePanel", 0) // 0: 列表, 1: 输入

	rego.UseKey(c, func(key rego.Key, r rune) {
		switch key {
		case rego.KeyTab:
			activePanel.Set((activePanel.Val + 1) % 2)
		}
		switch r {
		case '1':
			filter.Set(FilterAll)
		case '2':
			filter.Set(FilterActive)
		case '3':
			filter.Set(FilterCompleted)
		case 'q':
			c.Quit()
		}
	})

	// 过滤后的任务列表
	filteredTodos := rego.UseMemo(c, func() []Todo {
		result := make([]Todo, 0)
		for _, todo := range todos.Val {
			switch filter.Val {
			case FilterAll:
				result = append(result, todo)
			case FilterActive:
				if !todo.Completed {
					result = append(result, todo)
				}
			case FilterCompleted:
				if todo.Completed {
					result = append(result, todo)
				}
			}
		}
		return result
	}, todos.Val, filter.Val)

	// 统计
	activeCount := 0
	completedCount := 0
	for _, todo := range todos.Val {
		if todo.Completed {
			completedCount++
		} else {
			activeCount++
		}
	}

	return rego.VStack(
		// 顶部标题栏
		Header(c.Child("header")),

		rego.Text(""),

		// 过滤器栏
		FilterBar(c.Child("filter"), filter.Val, filter.Set),

		rego.Text(""),

		// 主体区域
		rego.HStack(
			// 左侧：任务列表
			TodoList(c.Child("list"), filteredTodos, todos, activePanel.Val == 0),

			rego.Text("  "),

			// 右侧：输入和统计
			rego.VStack(
				// 添加任务面板
				AddTodoPanel(c.Child("add"), todos, activePanel.Val == 1),

				rego.Text(""),

				// 统计面板
				StatsPanel(c.Child("stats"), len(todos.Val), activeCount, completedCount),
			).Flex(1),
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
			rego.Text("📝 Rego Todo List").Bold().Color(rego.Cyan),
			rego.Spacer(),
			rego.Text("[Tab] 切换面板").Dim(),
			rego.Text("  "),
			rego.Text("[q] 退出").Dim(),
		),
	).Border(rego.BorderDouble).BorderColor(rego.Cyan).Padding(0, 1)
}

// =============================================================================
// FilterBar 组件
// =============================================================================

func FilterBar(c rego.C, current int, setFilter func(int)) rego.Node {
	filters := []struct {
		label string
		value int
	}{
		{"全部", FilterAll},
		{"未完成", FilterActive},
		{"已完成", FilterCompleted},
	}

	return rego.Box(
		rego.HStack(
			rego.Text("过滤: ").Dim(),
			rego.For(filters, func(f struct {
				label string
				value int
			}, i int) rego.Node {
				isActive := f.value == current
				text := rego.Text(fmt.Sprintf(" [%d] %s ", i+1, f.label))

				if isActive {
					return text.Bold().Color(rego.Black).Background(rego.Cyan)
				}
				return text.Color(rego.Gray)
			}),
			rego.Spacer(),
			rego.Text("[1/2/3] 切换过滤").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(rego.Gray).Padding(0, 1)
}

// =============================================================================
// TodoList 组件
// =============================================================================

func TodoList(c rego.C, filteredTodos []Todo, allTodos *rego.State[[]Todo], active bool) rego.Node {
	selected := rego.Use(c, "selected", 0)

	// 处理键盘事件
	if active {
		rego.UseKey(c, func(key rego.Key, r rune) {
			switch key {
			case rego.KeyUp:
				if selected.Val > 0 {
					selected.Set(selected.Val - 1)
				}
			case rego.KeyDown:
				if selected.Val < len(filteredTodos)-1 {
					selected.Set(selected.Val + 1)
				}
			case rego.KeyEnter:
				// 切换完成状态
				if len(filteredTodos) > 0 && selected.Val < len(filteredTodos) {
					toggleTodo(allTodos, filteredTodos[selected.Val].Text)
				}
			}
			switch r {
			case 'd':
				// 删除任务
				if len(filteredTodos) > 0 && selected.Val < len(filteredTodos) {
					deleteTodo(allTodos, filteredTodos[selected.Val].Text)
					if selected.Val >= len(filteredTodos)-1 && selected.Val > 0 {
						selected.Set(selected.Val - 1)
					}
				}
			case 'x':
				// 清除已完成
				newTodos := make([]Todo, 0)
				for _, t := range allTodos.Val {
					if !t.Completed {
						newTodos = append(newTodos, t)
					}
				}
				allTodos.Set(newTodos)
				selected.Set(0)
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
				rego.Text("📋 任务列表").Bold(),
				rego.When(active, rego.Text(" ●").Color(rego.Green)),
			),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			// 任务列表
			rego.ScrollBox(c.Child("scroll"),
				rego.WhenElse(len(filteredTodos) == 0,
					rego.VStack(
						rego.Text(""),
						rego.Text("暂无任务").Dim(),
						rego.Text("切换到右侧面板添加新任务").Dim(),
					),
					rego.For(filteredTodos, func(todo Todo, i int) rego.Node {
						return TodoItem(c.Child("item", i), todo, i == selected.Val, active)
					}),
				),
			).Flex(1),

			rego.Spacer(),
			rego.Text("─────────────────────────────────").Dim(),
			rego.Text("[↑/↓] 选择  [Enter] 切换  [d] 删除").Dim(),
		),
	).Flex(2).Border(rego.BorderSingle).BorderColor(borderColor).Padding(1, 2)
}

// =============================================================================
// TodoItem 组件
// =============================================================================

func TodoItem(c rego.C, todo Todo, isSelected bool, panelActive bool) rego.Node {
	// 图标
	icon := "○"
	if todo.Completed {
		icon = "●"
	}

	// 前缀
	prefix := "  "
	if isSelected && panelActive {
		prefix = "▸ "
	}

	// 颜色
	textColor := rego.White
	iconColor := rego.Gray
	if todo.Completed {
		textColor = rego.Gray
		iconColor = rego.Green
	}
	if isSelected && panelActive {
		textColor = rego.Green
	}

	text := rego.Text(todo.Text).Color(textColor)
	if todo.Completed {
		text = text.Dim()
	}

	return rego.HStack(
		rego.Text(prefix).Color(rego.Green),
		rego.Text(icon).Color(iconColor),
		rego.Text(" "),
		text,
	)
}

// =============================================================================
// AddTodoPanel 组件
// =============================================================================

func AddTodoPanel(c rego.C, todos *rego.State[[]Todo], active bool) rego.Node {
	inputText := rego.Use(c, "input", "")
	focus := rego.UseFocus(c)

	// 处理输入
	if active {
		rego.UseKey(c, func(key rego.Key, r rune) {
			switch key {
			case rego.KeyEnter:
				if len(inputText.Val) > 0 {
					newTodo := Todo{Text: inputText.Val, Completed: false}
					todos.Set(append(todos.Val, newTodo))
					inputText.Set("")
				}
			case rego.KeyBackspace:
				if len(inputText.Val) > 0 {
					runes := []rune(inputText.Val)
					inputText.Set(string(runes[:len(runes)-1]))
				}
			case rego.KeyEsc:
				inputText.Set("")
			default:
				if r != 0 {
					inputText.Set(inputText.Val + string(r))
				}
			}
		})
	}

	borderColor := rego.Gray
	if active {
		borderColor = rego.Green
	}

	displayText := inputText.Val
	if displayText == "" {
		displayText = "输入新任务..."
	}

	return rego.Box(
		rego.VStack(
			rego.HStack(
				rego.Text("➕ 添加任务").Bold(),
				rego.When(active, rego.Text(" ●").Color(rego.Green)),
			),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			// 输入框
			rego.Box(
				rego.HStack(
					rego.Text("> ").Color(rego.Green),
					rego.WhenElse(inputText.Val == "",
						rego.Text(displayText).Dim(),
						rego.Text(inputText.Val).Color(rego.White),
					),
					rego.When(active && focus.IsFocused,
						rego.Text("▌").Color(rego.Green).Blink(),
					),
				),
			).Border(rego.BorderSingle).BorderColor(rego.If(active, rego.Cyan, rego.Gray)).Padding(0, 1),

			rego.Text(""),

			// 添加按钮
			rego.HStack(
				rego.Button(c.Child("btn-add"), rego.ButtonProps{
					Label:   " ✓ 添加 ",
					Primary: len(inputText.Val) > 0,
					OnClick: func() {
						if len(inputText.Val) > 0 {
							newTodo := Todo{Text: inputText.Val, Completed: false}
							todos.Set(append(todos.Val, newTodo))
							inputText.Set("")
						}
					},
				}),
				rego.Text(" "),
				rego.Button(c.Child("btn-clear"), rego.ButtonProps{
					Label: " ✕ 清空 ",
					OnClick: func() {
						inputText.Set("")
					},
				}),
			),

			rego.Spacer(),
			rego.Text("─────────────────────────").Dim(),
			rego.Text("[Enter] 添加  [Esc] 清空").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(borderColor).Padding(1, 2)
}

// =============================================================================
// StatsPanel 组件
// =============================================================================

func StatsPanel(c rego.C, total, active, completed int) rego.Node {
	// 计算完成百分比
	percent := 0
	if total > 0 {
		percent = (completed * 100) / total
	}

	return rego.Box(
		rego.VStack(
			rego.Text("📊 统计").Bold().Color(rego.Yellow),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			rego.HStack(
				rego.Text("总计: "),
				rego.Text(fmt.Sprintf("%d", total)).Bold().Color(rego.Cyan),
			),
			rego.HStack(
				rego.Text("未完成: "),
				rego.Text(fmt.Sprintf("%d", active)).Bold().Color(rego.Yellow),
			),
			rego.HStack(
				rego.Text("已完成: "),
				rego.Text(fmt.Sprintf("%d", completed)).Bold().Color(rego.Green),
			),

			rego.Text(""),

			// 进度条
			rego.Text("完成进度:"),
			ProgressBar(percent),

			rego.Spacer(),
			rego.Text("─────────────────────────").Dim(),
			rego.Text("[x] 清除已完成").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(rego.Yellow).Padding(1, 2).Flex(1)
}

// =============================================================================
// ProgressBar 组件
// =============================================================================

func ProgressBar(percent int) rego.Node {
	width := 20
	filled := (percent * width) / 100
	if filled > width {
		filled = width
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	color := rego.Red
	if percent >= 50 {
		color = rego.Yellow
	}
	if percent >= 80 {
		color = rego.Green
	}

	return rego.HStack(
		rego.Text("["),
		rego.Text(bar).Color(color),
		rego.Text("]"),
		rego.Text(fmt.Sprintf(" %d%%", percent)),
	)
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
			rego.Text("Rego Todo - 一个优雅的任务管理应用").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(rego.Gray).Padding(0, 1)
}

// =============================================================================
// 辅助函数
// =============================================================================

func toggleTodo(todos *rego.State[[]Todo], text string) {
	newTodos := make([]Todo, len(todos.Val))
	for i, t := range todos.Val {
		if t.Text == text {
			newTodos[i] = Todo{Text: t.Text, Completed: !t.Completed}
		} else {
			newTodos[i] = t
		}
	}
	todos.Set(newTodos)
}

func deleteTodo(todos *rego.State[[]Todo], text string) {
	newTodos := make([]Todo, 0)
	for _, t := range todos.Val {
		if t.Text != text {
			newTodos = append(newTodos, t)
		}
	}
	todos.Set(newTodos)
}

func main() {
	if err := rego.Run(App); err != nil {
		log.Fatal(err)
	}
}
