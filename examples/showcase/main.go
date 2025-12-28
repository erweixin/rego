package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	rego "github.com/erweixin/rego"
)

// =============================================================================
// Rego Showcase - 终极推广示例
// 设计目标：20-30秒完整展示 Rego 的所有核心魅力
// =============================================================================

func main() {
	if err := rego.Run(App); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

func App(c rego.C) rego.Node {
	activeTab := rego.Use(c, "activeTab", 0)
	tabNames := []string{"📊 Monitor", "🧩 Components", "🚀 Stream", "⚡ Hooks"}

	// 自动轮换（为 GIF 设计）
	rego.UseEffect(c, func() func() {
		ticker := time.NewTicker(6 * time.Second)
		go func() {
			for range ticker.C {
				activeTab.Update(func(v int) int { return (v + 1) % 4 })
			}
		}()
		return ticker.Stop
	})

	// 键盘交互
	rego.UseKey(c, func(key rego.Key, r rune) {
		switch r {
		case '1':
			activeTab.Set(0)
		case '2':
			activeTab.Set(1)
		case '3':
			activeTab.Set(2)
		case '4':
			activeTab.Set(3)
		case 'q':
			c.Quit()
		}
	})

	return rego.VStack(
		// 顶部：炫酷动态标题
		AnimatedHeader(c.Child("header")),

		rego.Text(""),

		// Tab 切换栏
		TabBar(c.Child("tabs"), tabNames, activeTab.Val),

		rego.Text(""),

		// 主内容区
		rego.VStack(
			renderTab(c, activeTab.Val),
		).Flex(1),

		rego.Text(""),

		// 底部状态栏
		FooterBar(c.Child("footer")),
	).Padding(1, 2)
}

func renderTab(c rego.C, activeTab int) rego.Node {
	switch activeTab {
	case 0:
		return MonitorView(c.Child("monitor"))
	case 1:
		return ComponentsView(c.Child("components"))
	case 2:
		return StreamView(c.Child("stream"))
	case 3:
		return HooksView(c.Child("hooks"))
	default:
		return MonitorView(c.Child("monitor"))
	}
}

// =============================================================================
// AnimatedHeader - 炫酷动态标题，带彩虹色和动画
// =============================================================================

func AnimatedHeader(c rego.C) rego.Node {
	tick := rego.Use(c, "tick", 0)
	now := rego.Use(c, "now", time.Now())

	rego.UseEffect(c, func() func() {
		ticker := time.NewTicker(150 * time.Millisecond)
		go func() {
			for range ticker.C {
				tick.Update(func(v int) int { return v + 1 })
				now.Set(time.Now())
			}
		}()
		return ticker.Stop
	})

	// 彩虹色 Logo
	logo := "R E G O"
	colors := []rego.Color{rego.Red, rego.Yellow, rego.Green, rego.Cyan, rego.Blue, rego.Magenta}

	return rego.Box(
		rego.HStack(
			// 动态彩虹 Logo
			rego.HStack(
				rego.For(strings.Split(logo, ""), func(ch string, i int) rego.Node {
					if ch == " " {
						return rego.Text(" ")
					}
					colorIdx := (tick.Val + i) % len(colors)
					return rego.Text(ch).Bold().Color(colors[colorIdx])
				}),
			),
			rego.Text("  "),
			rego.Text("React Hooks for Go TUI").Dim(),
			rego.Spacer(),
			// 动态 Spinner
			rego.Text(spinnerFrame(tick.Val)).Color(rego.Cyan),
			rego.Text(" "),
			rego.Text(now.Val.Format("15:04:05")).Color(rego.Yellow),
		),
	).Border(rego.BorderDouble).BorderColor(rego.Cyan).Padding(0, 2)
}

func spinnerFrame(tick int) string {
	frames := []string{"◐", "◓", "◑", "◒"}
	return frames[tick%len(frames)]
}

// =============================================================================
// TabBar - Tab 切换栏
// =============================================================================

func TabBar(c rego.C, tabs []string, active int) rego.Node {
	return rego.HStack(
		rego.For(tabs, func(tab string, i int) rego.Node {
			isActive := i == active
			text := rego.Text(fmt.Sprintf(" %s ", tab))
			if isActive {
				return text.Bold().Color(rego.Black).Background(rego.Cyan)
			}
			return text.Color(rego.Gray)
		}),
		rego.Spacer(),
		rego.Text("⟳ Auto-rotating").Dim().Italic(),
	)
}

// =============================================================================
// MonitorView - 系统监控面板（左右分栏）
// =============================================================================

func MonitorView(c rego.C) rego.Node {
	cpu := rego.Use(c, "cpu", 35)
	mem := rego.Use(c, "mem", 62)
	disk := rego.Use(c, "disk", 45)
	net := rego.Use(c, "net", 28)
	requests := rego.Use(c, "requests", 1024)
	uptime := rego.Use(c, "uptime", 0)
	logs := rego.Use(c, "logs", []string{
		"[INFO] System initialized",
		"[INFO] Rego engine started",
	})

	// 模拟实时数据
	rego.UseEffect(c, func() func() {
		ticker := time.NewTicker(400 * time.Millisecond)
		go func() {
			for range ticker.C {
				cpu.Update(func(v int) int { return clamp(v+rand.Intn(9)-4, 10, 95) })
				mem.Update(func(v int) int { return clamp(v+rand.Intn(5)-2, 30, 90) })
				disk.Update(func(v int) int { return clamp(v+rand.Intn(3)-1, 20, 80) })
				net.Update(func(v int) int { return rand.Intn(100) })
				requests.Update(func(v int) int { return v + rand.Intn(100) })
				uptime.Update(func(v int) int { return v + 1 })

				// 添加新日志
				if rand.Intn(3) == 0 {
					newLog := fmt.Sprintf("[%s] Event #%d processed",
						time.Now().Format("15:04:05"), rand.Intn(1000))
					logs.Update(func(l []string) []string {
						if len(l) >= 6 {
							l = l[1:]
						}
						return append(l, newLog)
					})
				}
			}
		}()
		return ticker.Stop
	})

	return rego.HStack(
		// 左侧：指标面板
		rego.Box(
			rego.VStack(
				rego.Text("📈 System Metrics").Bold().Color(rego.Green),
				rego.Divider().Color(rego.Gray),
				rego.Text(""),

				ProgressBar("CPU", cpu.Val, cpuColor(cpu.Val)),
				ProgressBar("MEM", mem.Val, memColor(mem.Val)),
				ProgressBar("DISK", disk.Val, rego.Blue),
				ProgressBar("NET", net.Val, rego.Magenta),

				rego.Text(""),

				rego.HStack(
					rego.Text("📡 "),
					rego.Text(fmt.Sprintf("%d", requests.Val)).Bold().Color(rego.Cyan),
					rego.Text(" req/s").Dim(),
					rego.Spacer(),
					rego.Text("⏱ "),
					rego.Text(formatUptime(uptime.Val)).Color(rego.Yellow),
				),

				rego.Spacer(),
				StatusIndicator(cpu.Val, mem.Val),
			),
		).Border(rego.BorderRounded).Padding(1, 2).Flex(1),

		rego.Text(" "),

		// 右侧：日志面板
		rego.Box(
			rego.VStack(
				rego.Text("📜 Live Logs").Bold().Color(rego.Yellow),
				rego.Divider().Color(rego.Gray),
				rego.Text(""),
				rego.For(logs.Val, func(log string, i int) rego.Node {
					return rego.Text(log).Dim()
				}),
				rego.Spacer(),
				rego.Text("Streaming...").Italic().Color(rego.Gray),
			),
		).Border(rego.BorderRounded).Padding(1, 2).Flex(1),
	).Flex(1)
}

// =============================================================================
// ComponentsView - 组件展示面板
// =============================================================================

func ComponentsView(c rego.C) rego.Node {
	checked1 := rego.Use(c, "checked1", true)
	checked2 := rego.Use(c, "checked2", false)
	inputVal := rego.Use(c, "inputVal", "Hello Rego!")
	btnClicks := rego.Use(c, "btnClicks", 0)

	return rego.HStack(
		// 左侧：表单组件
		rego.Box(
			rego.VStack(
				rego.Text("🧩 Form Components").Bold().Color(rego.Magenta),
				rego.Divider().Color(rego.Gray),
				rego.Text(""),

				// TextInput
				rego.Text("TextInput:").Dim(),
				rego.TextInput(c.Child("input"), rego.TextInputProps{
					Value:       inputVal.Val,
					Placeholder: "Type here...",
					OnChanged:   func(s string) { inputVal.Set(s) },
				}),

				rego.Text(""),

				// Checkboxes
				rego.Text("Checkbox:").Dim(),
				rego.Checkbox(c.Child("cb1"), rego.CheckboxProps{
					Label:     "Enable feature A",
					Checked:   checked1.Val,
					OnChanged: func(v bool) { checked1.Set(v) },
				}),
				rego.Checkbox(c.Child("cb2"), rego.CheckboxProps{
					Label:     "Enable feature B",
					Checked:   checked2.Val,
					OnChanged: func(v bool) { checked2.Set(v) },
				}),

				rego.Text(""),

				// Buttons
				rego.Text("Buttons:").Dim(),
				rego.HStack(
					rego.Button(c.Child("btn1"), rego.ButtonProps{
						Label:   "Primary",
						Primary: true,
						OnClick: func() { btnClicks.Update(func(v int) int { return v + 1 }) },
					}),
					rego.Text(" "),
					rego.Button(c.Child("btn2"), rego.ButtonProps{
						Label:   "Secondary",
						OnClick: func() {},
					}),
				),

				rego.Spacer(),
				rego.Text(fmt.Sprintf("Clicks: %d", btnClicks.Val)).Color(rego.Cyan),
			),
		).Border(rego.BorderRounded).Padding(1, 2).Flex(1),

		rego.Text(" "),

		// 右侧：其他组件
		rego.Box(
			rego.VStack(
				rego.Text("✨ More Components").Bold().Color(rego.Cyan),
				rego.Divider().Color(rego.Gray),
				rego.Text(""),

				// Spinner
				rego.Text("Spinner:").Dim(),
				rego.HStack(
					rego.Spinner(c.Child("spinner1"), "Loading data..."),
				),

				rego.Text(""),

				// Layout showcase
				rego.Text("Nested Boxes:").Dim(),
				rego.HStack(
					rego.Box(rego.Text("A").Color(rego.Red)).Border(rego.BorderSingle).Padding(0, 1),
					rego.Box(rego.Text("B").Color(rego.Green)).Border(rego.BorderSingle).Padding(0, 1),
					rego.Box(rego.Text("C").Color(rego.Blue)).Border(rego.BorderSingle).Padding(0, 1),
				).Gap(1),

				rego.Text(""),

				// Colors showcase
				rego.Text("Color Palette:").Dim(),
				ColorPalette(),

				rego.Spacer(),
				rego.Text("Use Tab to navigate").Italic().Color(rego.Gray),
			),
		).Border(rego.BorderRounded).Padding(1, 2).Flex(1),
	).Flex(1)
}

func ColorPalette() rego.Node {
	colors := []rego.Color{
		rego.Red, rego.Yellow, rego.Green, rego.Cyan, rego.Blue, rego.Magenta,
	}
	return rego.HStack(
		rego.For(colors, func(color rego.Color, i int) rego.Node {
			return rego.Text("██").Color(color)
		}),
	)
}

// =============================================================================
// StreamView - AI 流式输出展示
// =============================================================================

func StreamView(c rego.C) rego.Node {
	text := rego.Use(c, "streamText", "")
	charIndex := rego.Use(c, "charIndex", 0)
	isTyping := rego.Use(c, "isTyping", true)

	content := `## 🤖 AI Agent Streaming

Rego is **perfect** for building AI Agent CLIs:

### Key Features
1. **TailBox** - Auto-scrolling container
2. **Bridge** - Agent communication mechanism  
3. **Markdown** - Rich text rendering

### Code Example
` + "```go" + `
func AgentUI(c rego.C) rego.Node {
    bridge := rego.UseBridge[State, Q, A](c, init)
    
    rego.UseEffect(c, func() func() {
        go agent.Run(bridge.Handle())
        return nil
    })
    
    return rego.Markdown(bridge.State().Response)
}
` + "```" + `

> ✨ Build streaming UIs in **minutes**, not hours!

---
*This text is being streamed character by character...*`

	rego.UseEffect(c, func() func() {
		ticker := time.NewTicker(35 * time.Millisecond)
		go func() {
			for range ticker.C {
				runes := []rune(content)
				if charIndex.Val < len(runes) {
					charIndex.Update(func(v int) int { return v + 1 })
					text.Set(string(runes[:charIndex.Val]))
					isTyping.Set(true)
				} else {
					isTyping.Set(false)
					time.Sleep(3 * time.Second)
					charIndex.Set(0)
					text.Set("")
				}
			}
		}()
		return ticker.Stop
	})

	return rego.Box(
		rego.VStack(
			rego.HStack(
				rego.Text("🚀 Streaming Demo").Bold().Color(rego.Magenta),
				rego.Spacer(),
				rego.WhenElse(isTyping.Val,
					rego.Text("● typing...").Color(rego.Green),
					rego.Text("○ paused").Color(rego.Gray),
				),
			),
			rego.Divider().Color(rego.Gray),

			rego.TailBox(c.Child("stream-scroll"),
				rego.Box(
					rego.VStack(
						rego.Markdown(text.Val+rego.If(isTyping.Val, "▍", "")),
					),
				).Padding(1, 1),
			).Flex(1),
		),
	).Border(rego.BorderRounded).Padding(1, 2).Flex(1)
}

// =============================================================================
// HooksView - Hooks 演示面板
// =============================================================================

func HooksView(c rego.C) rego.Node {
	count := rego.Use(c, "count", 0)
	history := rego.Use(c, "history", []int{0, 0, 0, 0, 0, 0, 0, 0})

	// 自动计数和历史记录
	rego.UseEffect(c, func() func() {
		ticker := time.NewTicker(300 * time.Millisecond)
		go func() {
			for range ticker.C {
				count.Update(func(v int) int { return v + 1 })
				history.Update(func(h []int) []int {
					newH := append(h[1:], count.Val%20)
					return newH
				})
			}
		}()
		return ticker.Stop
	})

	// UseMemo 派生状态
	squared := rego.UseMemo(c, func() int {
		return count.Val * count.Val
	}, count.Val)

	doubled := rego.UseMemo(c, func() int {
		return count.Val * 2
	}, count.Val)

	return rego.HStack(
		// 左侧：Hook 演示
		rego.Box(
			rego.VStack(
				rego.Text("⚡ Hooks in Action").Bold().Color(rego.Yellow),
				rego.Divider().Color(rego.Gray),
				rego.Text(""),

				// Use
				rego.Text("rego.Use() - State Management").Dim(),
				rego.HStack(
					rego.Text("count = "),
					rego.Text(fmt.Sprintf("%d", count.Val)).Bold().Color(rego.Cyan),
				),

				rego.Text(""),

				// UseMemo
				rego.Text("rego.UseMemo() - Computed Values").Dim(),
				rego.HStack(
					rego.Text("squared = "),
					rego.Text(fmt.Sprintf("%d", squared)).Bold().Color(rego.Magenta),
					rego.Text("  doubled = "),
					rego.Text(fmt.Sprintf("%d", doubled)).Bold().Color(rego.Green),
				),

				rego.Text(""),

				// 可视化
				rego.Text("Visual:").Dim(),
				rego.Text(waveVisualization(count.Val)).Color(rego.Cyan),

				rego.Spacer(),

				// 状态图标
				rego.HStack(
					rego.WhenElse(count.Val%2 == 0,
						rego.Text("◉ EVEN").Color(rego.Green),
						rego.Text("◎ ODD").Color(rego.Red),
					),
					rego.Spacer(),
					rego.WhenElse(count.Val%5 == 0,
						rego.Text("★ FIVE").Color(rego.Yellow),
						rego.Text("☆").Color(rego.Gray),
					),
				),
			),
		).Border(rego.BorderRounded).Padding(1, 2).Flex(1),

		rego.Text(" "),

		// 右侧：迷你图表
		rego.Box(
			rego.VStack(
				rego.Text("📊 Live Chart").Bold().Color(rego.Blue),
				rego.Divider().Color(rego.Gray),
				rego.Text(""),

				// 迷你柱状图
				rego.Text("UseEffect() - Side Effects").Dim(),
				rego.Text(""),
				MiniBarChart(history.Val),

				rego.Spacer(),

				rego.Text("// Data updates every 300ms").Italic().Color(rego.Gray),
			),
		).Border(rego.BorderRounded).Padding(1, 2).Flex(1),
	).Flex(1)
}

func MiniBarChart(data []int) rego.Node {
	bars := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

	return rego.VStack(
		rego.For([]int{0}, func(_ int, _ int) rego.Node {
			result := ""
			for _, v := range data {
				idx := v * len(bars) / 20
				if idx >= len(bars) {
					idx = len(bars) - 1
				}
				result += bars[idx] + " "
			}
			return rego.Text(result).Color(rego.Cyan)
		}),
	)
}

func waveVisualization(n int) string {
	wave := ""
	for i := 0; i < 30; i++ {
		angle := float64(n+i) * 0.3
		height := int(math.Sin(angle)*3 + 3)
		chars := []string{" ", "▁", "▂", "▃", "▄", "▅", "▆"}
		if height >= 0 && height < len(chars) {
			wave += chars[height]
		} else {
			wave += "▃"
		}
	}
	return wave
}

// =============================================================================
// FooterBar - 底部状态栏
// =============================================================================

func FooterBar(c rego.C) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("[1-4]").Bold().Color(rego.Cyan),
			rego.Text(" Switch").Dim(),
			rego.Text("  "),
			rego.Text("[Tab]").Bold().Color(rego.Green),
			rego.Text(" Focus").Dim(),
			rego.Text("  "),
			rego.Text("[q]").Bold().Color(rego.Red),
			rego.Text(" Quit").Dim(),
			rego.Spacer(),
			rego.Text("⭐ github.com/erweixin/rego").Italic().Color(rego.Yellow),
		),
	).Border(rego.BorderSingle).Padding(0, 2)
}

// =============================================================================
// 辅助组件和函数
// =============================================================================

func ProgressBar(label string, percent int, color rego.Color) rego.Node {
	width := 20
	filled := (percent * width) / 100
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return rego.HStack(
		rego.Text(fmt.Sprintf("%-4s", label)).Bold(),
		rego.Text("[").Dim(),
		rego.Text(bar).Color(color),
		rego.Text("]").Dim(),
		rego.Text(fmt.Sprintf(" %3d%%", percent)),
	)
}

func StatusIndicator(cpu, mem int) rego.Node {
	status := "● All systems operational"
	color := rego.Green

	if cpu >= 80 || mem >= 85 {
		status = "● High load detected"
		color = rego.Red
	} else if cpu >= 60 || mem >= 70 {
		status = "● Moderate load"
		color = rego.Yellow
	}

	return rego.Text(status).Color(color)
}

func formatUptime(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func cpuColor(v int) rego.Color {
	if v >= 80 {
		return rego.Red
	}
	if v >= 50 {
		return rego.Yellow
	}
	return rego.Green
}

func memColor(v int) rego.Color {
	if v >= 85 {
		return rego.Red
	}
	if v >= 60 {
		return rego.Yellow
	}
	return rego.Green
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
