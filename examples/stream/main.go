package main

import (
	"fmt"
	"time"

	rego "github.com/erweixin/rego"
)

// =============================================================================
// StreamDemo - 一个纯粹展示流式输出和智能滚动的示例
// =============================================================================

func main() {
	if err := rego.Run(App); err != nil {
		panic(err)
	}
}

func App(c rego.C) rego.Node {
	messages := rego.Use(c, "messages", []string{})
	currentStream := rego.Use(c, "currentStream", "")
	isStreaming := rego.Use(c, "isStreaming", false)

	// 自动开始第一个流式任务
	rego.UseEffect(c, func() func() {
		if !isStreaming.Val && len(messages.Val) == 0 {
			startDemoStream(c, messages, currentStream, isStreaming)
		}
		return nil
	}, isStreaming.Val)

	rego.UseKey(c, func(key rego.Key, r rune) {
		if r == 'r' && !isStreaming.Val {
			// 按 R 重置并重新开始
			messages.Set([]string{})
			currentStream.Set("")
			startDemoStream(c, messages, currentStream, isStreaming)
		}
		if key == rego.KeyCtrlC {
			c.Quit()
		}
	})

	return rego.VStack(
		// 顶部标题栏
		rego.Box(
			rego.HStack(
				rego.Text("🚀 REGO STREAMING DEMO").Bold().Color(rego.Cyan),
				rego.Spacer(),
				rego.Stats(c.Child("stats")),
				rego.Text(" "),
				rego.Text(fmt.Sprintf("状态: %s", If(isStreaming.Val, "正在生成...", "就绪"))).
					Color(If(isStreaming.Val, rego.Yellow, rego.Green)),
			),
		).Border(rego.BorderSingle).Padding(0, 1),

		rego.Text(""),

		// 主视图区域：展示 StreamView 的核心逻辑
		rego.TailBox(c.Child("chat-scroll"),
			rego.Box(
				rego.VStack(
					// 已完成的消息历史
					rego.For(messages.Val, func(msg string, i int) rego.Node {
						return rego.VStack(
							rego.Text(fmt.Sprintf("--- 历史消息 #%d ---", i+1)).Dim(),
							rego.Markdown(msg),
							rego.Text(""),
						)
					}),

					// 当前正在流出的消息
					rego.When(isStreaming.Val || currentStream.Val != "",
						rego.VStack(
							rego.Text("--- AI 正在输入 ---").Color(rego.Yellow).Italic(),
							rego.Markdown(currentStream.Val+"▍"),
						),
					),
				),
			).Apply(rego.NewStyle().Padding(1, 2)),
		).Flex(1),

		rego.Text(""),

		// 底部操作提示
		rego.HStack(
			rego.Text(" [R] 重新运行示例 ").Background(rego.Blue).Color(rego.White),
			rego.Text("  "),
			rego.Text(" [Ctrl+C] 退出 ").Background(rego.Gray).Color(rego.White),
			rego.Spacer(),
			rego.Text("提示：试着在生成时向上滚动鼠标，跟随会自动停止。滚回底部则恢复。").Dim(),
		),
	).Padding(1, 2)
}

// startDemoStream 启动一个模拟的长文本流
func startDemoStream(c rego.C, history *rego.State[[]string], current *rego.State[string], status *rego.State[bool]) {
	status.Set(true)
	current.Set("")

	go func() {
		content := `
## 正在演示智能滚动 (Auto-Tail)

当内容在流式增长时，` + "`TailBox`" + ` 会确保你的视口始终跟随最新的 Token。

### 为什么这很重要？
1. **无需手动滚动**：Agent 输出非常快，手动滚动太累。
2. **不闪烁**：即使 Markdown 的高度在不断变化，Rego 也能保持稳定。

### 复杂内容测试
下面是一段带高亮的代码块，观察它增加行数时视口的表现：

` + "```go" + `
package main

import "fmt"

func demonstrate() {
    for i := 0; i < 10; i++ {
        fmt.Printf("Token sequence: %d\n", i)
        // 这里的代码块会不断变长
    }
}
` + "```" + `

### 列表增长
- 自动生成项 1
- 自动生成项 2
- 自动生成项 3
- 自动生成项 4
- 自动生成项 5
- 自动生成项 6
- 自动生成项 7
- 自动生成项 8
- 自动生成项 9
- 自动生成项 10

### 总结
这就是 Rego 的 ` + "`StreamView`" + ` 组件背后的逻辑。
它让 Agent 开发体验 (DX) 达到了工业级水准。
`
		// 模拟 Token 逐个流出
		fullRunes := []rune(content)
		currentText := ""
		for _, r := range fullRunes {
			currentText += string(r)
			current.Set(currentText)
			// 模拟随机的 Token 产生速度
			time.Sleep(20 * time.Millisecond)
		}

		// 完成流，存入历史
		time.Sleep(500 * time.Millisecond)
		history.Update(func(h []string) []string {
			return append(h, currentText)
		})
		current.Set("")
		status.Set(false)
		c.Refresh()
	}()
}

// 简单的辅助函数
func If[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}
