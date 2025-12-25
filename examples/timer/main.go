package main

import (
	"fmt"
	"time"

	rego "github.com/erweixin/rego"
)

// =============================================================================
// Timer 示例 - 展示 UseEffect、UseMemo 和精美的 UI 布局
// =============================================================================

func App(c rego.C) rego.Node {
	activeTab := rego.Use(c, "activeTab", 0) // 0: 秒表, 1: 倒计时

	rego.UseKey(c, func(key rego.Key, r rune) {
		switch key {
		case rego.KeyTab:
			activeTab.Set((activeTab.Val + 1) % 2)
		}
		switch r {
		case '1':
			activeTab.Set(0)
		case '2':
			activeTab.Set(1)
		case 'q':
			c.Quit()
		}
	})

	return rego.VStack(
		// 顶部标题栏
		Header(c.Child("header")),

		rego.Text(""),

		// Tab 切换栏
		TabBar(c.Child("tabs"), activeTab.Val, activeTab.Set),

		rego.Text(""),

		// 主体内容
		rego.WhenElse(activeTab.Val == 0,
			StopwatchPanel(c.Child("stopwatch")),
			CountdownPanel(c.Child("countdown")),
		),

		rego.Spacer(),

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
			rego.Text("⏱️ Rego Timer").Bold().Color(rego.Cyan),
			rego.Spacer(),
			rego.Text(time.Now().Format("2006-01-02")).Dim(),
		),
	).Border(rego.BorderDouble).BorderColor(rego.Cyan).Padding(0, 1)
}

// =============================================================================
// TabBar 组件
// =============================================================================

func TabBar(c rego.C, activeIndex int, setActive func(int)) rego.Node {
	tabs := []string{"⏱️ 秒表", "⏳ 倒计时"}

	return rego.Box(
		rego.HStack(
			rego.For(tabs, func(tab string, i int) rego.Node {
				isActive := i == activeIndex
				style := rego.Text(" " + tab + " ")

				if isActive {
					return style.Bold().Color(rego.Black).Background(rego.Cyan)
				}
				return style.Color(rego.Gray)
			}),
			rego.Spacer(),
			rego.Text("[Tab] 或 [1/2] 切换").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(rego.Gray).Padding(0, 1)
}

// =============================================================================
// StopwatchPanel 组件 - 秒表功能
// =============================================================================

func StopwatchPanel(c rego.C) rego.Node {
	seconds := rego.Use(c, "seconds", 0)
	running := rego.Use(c, "running", false)
	laps := rego.Use(c, "laps", []int{})

	// 格式化时间
	formattedTime := rego.UseMemo(c, func() string {
		h := seconds.Val / 3600
		m := (seconds.Val % 3600) / 60
		s := seconds.Val % 60
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}, seconds.Val)

	// UseEffect：创建定时器
	rego.UseEffect(c, func() func() {
		if !running.Val {
			return nil
		}

		ticker := time.NewTicker(time.Second)
		done := make(chan bool)

		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					seconds.Update(func(s int) int { return s + 1 })
				}
			}
		}()

		return func() {
			ticker.Stop()
			close(done)
		}
	}, running.Val)

	// 键盘事件
	rego.UseKey(c, func(key rego.Key, r rune) {
		switch r {
		case ' ':
			running.Set(!running.Val)
		case 'r':
			seconds.Set(0)
			laps.Set([]int{})
		case 'l':
			if running.Val {
				laps.Set(append(laps.Val, seconds.Val))
			}
		}
	})

	// 状态文本
	statusText := "▶ 运行中"
	statusColor := rego.Green
	if !running.Val {
		statusText = "⏸ 已暂停"
		statusColor = rego.Yellow
	}

	return rego.HStack(
		// 左侧：主时间显示
		rego.Box(
			rego.VStack(
				rego.Text("秒表模式").Bold().Color(rego.Cyan),
				rego.Divider().Color(rego.Gray),
				rego.Text(""),

				// 大号时间显示
				rego.Box(
					rego.Text(formattedTime).Bold().Color(rego.White),
				).Border(rego.BorderRounded).BorderColor(rego.Cyan).Padding(2, 6),

				rego.Text(""),
				rego.Text(statusText).Color(statusColor),
				rego.Text(""),

				// 控制按钮
				rego.HStack(
					rego.Button(c.Child("btn-start"), rego.ButtonProps{
						Label:   rego.If(running.Val, " ⏸ 暂停 ", " ▶ 开始 "),
						Primary: !running.Val,
						OnClick: func() { running.Set(!running.Val) },
					}),
					rego.Text(" "),
					rego.Button(c.Child("btn-lap"), rego.ButtonProps{
						Label:   " 📍 计圈 ",
						OnClick: func() {
							if running.Val {
								laps.Set(append(laps.Val, seconds.Val))
							}
						},
					}),
					rego.Text(" "),
					rego.Button(c.Child("btn-reset"), rego.ButtonProps{
						Label: " ↺ 重置 ",
						OnClick: func() {
							seconds.Set(0)
							laps.Set([]int{})
						},
					}),
				),

				rego.Spacer(),
				rego.Text("─────────────────────────────").Dim(),
				rego.Text("[Space] 开始/暂停  [l] 计圈  [r] 重置").Dim(),
			),
		).Border(rego.BorderSingle).Padding(1, 2).Flex(2),

		rego.Text("  "),

		// 右侧：计圈记录
		rego.Box(
			rego.VStack(
				rego.Text("📍 计圈记录").Bold().Color(rego.Yellow),
				rego.Divider().Color(rego.Gray),
				rego.Text(""),

				rego.ScrollBox(c.Child("laps-scroll"),
					rego.WhenElse(len(laps.Val) == 0,
						rego.Text("暂无记录").Dim(),
						rego.For(laps.Val, func(lap int, i int) rego.Node {
							h := lap / 3600
							m := (lap % 3600) / 60
							s := lap % 60
							return rego.HStack(
								rego.Text(fmt.Sprintf("#%02d", i+1)).Color(rego.Gray),
								rego.Spacer(),
								rego.Text(fmt.Sprintf("%02d:%02d:%02d", h, m, s)).Color(rego.White),
							)
						}),
					),
				).Flex(1),

				rego.Spacer(),
				rego.Text(fmt.Sprintf("共 %d 圈", len(laps.Val))).Dim(),
			),
		).Border(rego.BorderSingle).Padding(1, 2).Flex(1),
	).Flex(1)
}

// =============================================================================
// CountdownPanel 组件 - 倒计时功能
// =============================================================================

func CountdownPanel(c rego.C) rego.Node {
	totalSeconds := rego.Use(c, "total", 300) // 默认 5 分钟
	remaining := rego.Use(c, "remaining", 300)
	running := rego.Use(c, "running", false)
	finished := rego.Use(c, "finished", false)

	// 预设时间
	presets := []struct {
		label   string
		seconds int
	}{
		{"1分钟", 60},
		{"5分钟", 300},
		{"10分钟", 600},
		{"30分钟", 1800},
	}

	// 格式化时间
	formattedTime := rego.UseMemo(c, func() string {
		m := remaining.Val / 60
		s := remaining.Val % 60
		return fmt.Sprintf("%02d:%02d", m, s)
	}, remaining.Val)

	// 计算进度百分比
	progress := rego.UseMemo(c, func() int {
		if totalSeconds.Val == 0 {
			return 0
		}
		return (remaining.Val * 100) / totalSeconds.Val
	}, remaining.Val, totalSeconds.Val)

	// UseEffect：倒计时逻辑
	rego.UseEffect(c, func() func() {
		if !running.Val || remaining.Val <= 0 {
			return nil
		}

		ticker := time.NewTicker(time.Second)
		done := make(chan bool)

		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					remaining.Update(func(s int) int {
						if s <= 1 {
							running.Set(false)
							finished.Set(true)
							return 0
						}
						return s - 1
					})
				}
			}
		}()

		return func() {
			ticker.Stop()
			close(done)
		}
	}, running.Val)

	// 键盘事件
	rego.UseKey(c, func(key rego.Key, r rune) {
		switch r {
		case ' ':
			if remaining.Val > 0 {
				running.Set(!running.Val)
				finished.Set(false)
			}
		case 'r':
			remaining.Set(totalSeconds.Val)
			running.Set(false)
			finished.Set(false)
		}
		switch key {
		case rego.KeyUp:
			newTotal := totalSeconds.Val + 60
			totalSeconds.Set(newTotal)
			if !running.Val {
				remaining.Set(newTotal)
			}
		case rego.KeyDown:
			if totalSeconds.Val > 60 {
				newTotal := totalSeconds.Val - 60
				totalSeconds.Set(newTotal)
				if !running.Val {
					remaining.Set(newTotal)
				}
			}
		}
	})

	// 状态和颜色
	displayColor := rego.White
	if finished.Val {
		displayColor = rego.Red
	} else if progress < 20 {
		displayColor = rego.Yellow
	}

	return rego.Box(
		rego.VStack(
			rego.Text("倒计时模式").Bold().Color(rego.Magenta),
			rego.Divider().Color(rego.Gray),
			rego.Text(""),

			// 大号时间显示
			rego.Center(
				rego.Box(
					rego.VStack(
						rego.When(finished.Val,
							rego.Text("🔔 时间到！").Bold().Color(rego.Red).Blink(),
						),
						rego.Text(formattedTime).Bold().Color(displayColor),
					),
				).Border(rego.BorderRounded).BorderColor(displayColor).Padding(2, 8),
			),

			rego.Text(""),

			// 进度条
			ProgressBar(c.Child("progress"), progress),

			rego.Text(""),

			// 预设按钮
			rego.HStack(
				rego.Text("预设: ").Dim(),
				rego.For(presets, func(p struct {
					label   string
					seconds int
				}, i int) rego.Node {
					return rego.HStack(
						rego.Button(c.Child("preset", i), rego.ButtonProps{
							Label:   p.label,
							Primary: totalSeconds.Val == p.seconds,
							OnClick: func() {
								totalSeconds.Set(p.seconds)
								remaining.Set(p.seconds)
								running.Set(false)
								finished.Set(false)
							},
						}),
						rego.Text(" "),
					)
				}),
			),

			rego.Text(""),

			// 控制按钮
			rego.HStack(
				rego.Button(c.Child("btn-start"), rego.ButtonProps{
					Label:   rego.If(running.Val, " ⏸ 暂停 ", " ▶ 开始 "),
					Primary: !running.Val && remaining.Val > 0,
					OnClick: func() {
						if remaining.Val > 0 {
							running.Set(!running.Val)
							finished.Set(false)
						}
					},
				}),
				rego.Text(" "),
				rego.Button(c.Child("btn-reset"), rego.ButtonProps{
					Label: " ↺ 重置 ",
					OnClick: func() {
						remaining.Set(totalSeconds.Val)
						running.Set(false)
						finished.Set(false)
					},
				}),
			),

			rego.Spacer(),
			rego.Text("─────────────────────────────────────────").Dim(),
			rego.Text("[Space] 开始/暂停  [↑/↓] 调整时间  [r] 重置").Dim(),
		),
	).Border(rego.BorderSingle).Padding(1, 2).Flex(1)
}

// =============================================================================
// ProgressBar 组件
// =============================================================================

func ProgressBar(c rego.C, percent int) rego.Node {
	width := 40
	filled := (percent * width) / 100
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}

	bar := ""
	for i := 0; i < width; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}

	color := rego.Green
	if percent < 50 {
		color = rego.Yellow
	}
	if percent < 20 {
		color = rego.Red
	}

	return rego.HStack(
		rego.Text("["),
		rego.Text(bar).Color(color),
		rego.Text("]"),
		rego.Text(fmt.Sprintf(" %3d%%", percent)),
	)
}

// =============================================================================
// Footer 组件
// =============================================================================

func Footer(c rego.C) rego.Node {
	return rego.Box(
		rego.HStack(
			rego.Text("Rego Timer").Dim(),
			rego.Spacer(),
			rego.Text("[q] 退出").Dim(),
		),
	).Border(rego.BorderSingle).BorderColor(rego.Gray).Padding(0, 1)
}

func main() {
	if err := rego.Run(App); err != nil {
		panic(err)
	}
}
