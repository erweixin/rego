package main

import (
	"strings"
	"time"
	"unicode"

	rego "github.com/erweixin/rego"
)

// =============================================================================
// App - 主应用
// =============================================================================

func App(c rego.C) rego.Node {
	activePanel := rego.Use(c, "activePanel", 0)
	messages := rego.Use(c, "messages", []Message{
		{Role: "system", Content: "欢迎使用 Rego Agent CLI! 这是一个多面板 Agent 界面示例。"},
	})
	inputText := rego.Use(c, "inputText", "")
	isThinking := rego.Use(c, "isThinking", false)
	streamingText := rego.Use(c, "streamingText", "")

	// 处理键盘事件
	rego.UseKey(c, func(key rego.Key, r rune) {
		switch {
		case key == rego.KeyTab:
			activePanel.Update(func(p int) int { return (p + 1) % 3 })
		case r == '1':
			activePanel.Set(0)
		case r == '2':
			activePanel.Set(1)
		case r == '3':
			activePanel.Set(2)
		case key == rego.KeyCtrlC:
			c.Quit()
		}
	})

	return rego.VStack(
		// Header
		Header(c.Child("header")),

		// Main content (3 panels)
		rego.HStack(
			// Left: Chat history
			ChatPanel(c.Child("chat"), messages.Val, activePanel.Val == 0),

			// Center: Input & thinking
			InputPanel(c.Child("input"), inputText, isThinking.Val, streamingText.Val, activePanel.Val == 1, func(text string) {
				// 发送消息
				newMsg := Message{Role: "user", Content: text}
				messages.Set(append(messages.Val, newMsg))
				inputText.Set("")
				isThinking.Set(true)

				// 模拟 AI 响应
				go simulateResponse(c, messages, isThinking, streamingText)
			}),

			// Right: Context/Files
			ContextPanel(c.Child("context"), activePanel.Val == 2),
		).Flex(1),

		// Footer
		Footer(c.Child("footer")),
	)
}

// =============================================================================
// Message 类型
// =============================================================================

type Message struct {
	Role    string // "user", "assistant", "system"
	Content string
}

// =============================================================================
// Header 组件
// =============================================================================

func Header(c rego.C) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("🤖 Rego Agent CLI").Bold().Color(rego.Cyan),
			rego.Spacer(),
			rego.Text("v0.1.0").Dim(),
		),
	).Border(rego.BorderSingle).Padding(0, 1)
}

// =============================================================================
// ChatPanel 组件 - 聊天历史
// =============================================================================

func ChatPanel(c rego.C, messages []Message, active bool) rego.Node {
	// 最多显示最近 10 条
	start := 0
	if len(messages) > 10 {
		start = len(messages) - 10
	}
	displayMessages := messages[start:]

	borderColor := rego.Gray
	if active {
		borderColor = rego.Green
	}

	return rego.Box(
		rego.VStack(
			rego.HStack(
				rego.Text("💬 对话历史").Bold(),
				rego.When(active, rego.Text(" [活动]").Color(rego.Green)),
			),
			rego.Text(strings.Repeat("─", 30)),
			rego.For(displayMessages, func(msg Message, i int) rego.Node {
				return MessageItem(c.Child("msg", i), msg)
			}),
			rego.Spacer(),
		),
	).Width(35).Border(rego.BorderSingle).BorderColor(borderColor).Padding(1, 1).Flex(1)
}

// MessageItem 消息项
func MessageItem(c rego.C, msg Message) rego.Node {
	var prefix string
	var color rego.Color

	switch msg.Role {
	case "user":
		prefix = "👤 "
		color = rego.Blue
	case "assistant":
		prefix = "🤖 "
		color = rego.Cyan
	case "system":
		prefix = "💡 "
		color = rego.Yellow
	}

	// 不再截断消息，使用 Markdown 渲染
	content := msg.Content

	return rego.VStack(
		rego.Text(prefix).Bold().Color(color),
		rego.Box(
			rego.Markdown(content),
		).Padding(0, 1),
	)
}

// =============================================================================
// InputPanel 组件 - 输入区域
// =============================================================================

func InputPanel(c rego.C, inputText *rego.State[string], thinking bool, streamingText string, active bool, onSubmit func(string)) rego.Node {
	borderColor := rego.Gray
	if active {
		borderColor = rego.Green
	}

	// 只在活动时处理输入
	rego.UseKey(c, func(key rego.Key, r rune) {
		if !active {
			return
		}

		switch {
		case key == rego.KeyEnter:
			if len(inputText.Val) > 0 && !thinking {
				onSubmit(inputText.Val)
			}
		case key == rego.KeyBackspace:
			if len(inputText.Val) > 0 {
				inputText.Set(inputText.Val[:len(inputText.Val)-1])
			}
		case unicode.IsPrint(r): // 可打印字符（包括中文）
			inputText.Set(inputText.Val + string(r))
		}
	})

	return rego.Box(
		rego.VStack(
			rego.HStack(
				rego.Text("📝 输入").Bold(),
				rego.When(active, rego.Text(" [活动]").Color(rego.Green)),
			),
			rego.Text(strings.Repeat("─", 30)),
			rego.Text(""),

			// 思考/流式输出区域
			rego.When(thinking,
				rego.VStack(
					rego.Text("🔄 思考中...").Color(rego.Yellow),
					rego.When(len(streamingText) > 0,
						rego.Text(streamingText).Color(rego.Cyan),
					),
				),
			),

			rego.Spacer(),

			// 输入框
			rego.Text("─────────────────────────────"),
			rego.HStack(
				rego.Text("> ").Color(rego.Green),
				rego.Text(inputText.Val).Color(rego.White),
				rego.WhenElse(active,
					rego.HStack(rego.Cursor(c), rego.Text("▌").Color(rego.White)),
					rego.Empty(),
				),
			),
			rego.Text(""),
			rego.Text("[Enter] 发送  [Tab] 切换面板").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(borderColor).Padding(1, 1).Flex(2)
}

// =============================================================================
// ContextPanel 组件 - 上下文/文件
// =============================================================================

func ContextPanel(c rego.C, active bool) rego.Node {
	files := []string{
		"📁 src/",
		"  📄 main.go",
		"  📄 app.go",
		"  📄 utils.go",
		"📁 docs/",
		"  📄 README.md",
		"📁 tests/",
	}

	borderColor := rego.Gray
	if active {
		borderColor = rego.Green
	}

	return rego.Box(
		rego.VStack(
			rego.HStack(
				rego.Text("📂 上下文").Bold(),
				rego.When(active, rego.Text(" [活动]").Color(rego.Green)),
			),
			rego.Text(strings.Repeat("─", 20)),
			rego.For(files, func(file string, i int) rego.Node {
				return rego.Text(file)
			}),
			rego.Spacer(),
			rego.Text("─────────────────"),
			rego.Text("工作目录:").Dim(),
			rego.Text("/project").Color(rego.Cyan),
		),
	).Width(25).Border(rego.BorderSingle).BorderColor(borderColor).Padding(1, 1).Flex(1)
}

// =============================================================================
// Footer 组件
// =============================================================================

func Footer(c rego.C) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("[1] 对话").Dim(),
			rego.Text("  "),
			rego.Text("[2] 输入").Dim(),
			rego.Text("  "),
			rego.Text("[3] 上下文").Dim(),
			rego.Spacer(),
			rego.Text("[Tab] 切换  [Ctrl+C] 退出").Dim(),
		),
	).Border(rego.BorderSingle).Padding(0, 1)
}

// =============================================================================
// 模拟 AI 响应
// =============================================================================

func simulateResponse(c rego.C, messages *rego.State[[]Message], isThinking *rego.State[bool], streamingText *rego.State[string]) {
	// 模拟思考延迟
	time.Sleep(500 * time.Millisecond)

	// 模拟流式输出
	response := "收到您的消息！\n\n### Rego 框架特点\n- **Hooks 风格**: 熟悉的状态管理\n- **声明式 UI**: 简单直观的布局\n\n```go\nfunc Hello(c rego.C) rego.Node {\n    return rego.Text(\"Hello Markdown!\")\n}\n```\n\n构建这类复杂 TUI 变得非常简单！"

	for i := range response {
		streamingText.Set(response[:i+1])
		time.Sleep(30 * time.Millisecond)
	}

	// 完成响应
	time.Sleep(200 * time.Millisecond)

	messages.Set(append(messages.Val, Message{
		Role:    "assistant",
		Content: response,
	}))
	isThinking.Set(false)
	streamingText.Set("")
}

func main() {
	if err := rego.Run(App); err != nil {
		panic(err)
	}
}
