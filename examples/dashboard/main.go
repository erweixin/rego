package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	rego "github.com/erweixin/rego"
)

// =============================================================================
// 自定义 Hooks - 展示代码复用和组合的优雅
// =============================================================================

// useInterval 封装定时任务
func useInterval(c rego.C, callback func(), delay time.Duration) {
	rego.UseEffect(c, func() func() {
		ticker := time.NewTicker(delay)
		go func() {
			for range ticker.C {
				callback()
			}
		}()
		return ticker.Stop
	})
}

// useSystemStats 模拟系统数据抓取
func useSystemStats(c rego.C) (cpu int, mem int, net int) {
	cpuState := rego.Use(c, "cpu", 20)
	memState := rego.Use(c, "mem", 45)
	netState := rego.Use(c, "net", 10)

	useInterval(c, func() {
		cpuState.Update(func(v int) int {
			delta := rand.Intn(11) - 5 // -5 to +5
			newVal := v + delta
			if newVal < 0 {
				newVal = 0
			}
			if newVal > 100 {
				newVal = 100
			}
			return newVal
		})
		memState.Update(func(v int) int {
			delta := rand.Intn(5) - 2 // -2 to +2
			newVal := v + delta
			if newVal < 0 {
				newVal = 0
			}
			if newVal > 100 {
				newVal = 100
			}
			return newVal
		})
		netState.Update(func(v int) int {
			return rand.Intn(100)
		})
		c.Refresh()
	}, 800*time.Millisecond)

	return cpuState.Val, memState.Val, netState.Val
}

// =============================================================================
// 酷炫组件 - 展示声明式 UI 的力量
// =============================================================================

// ProgressBar 进度条组件
func ProgressBar(label string, percent int, color rego.Color) rego.Node {
	width := 20
	filled := (percent * width) / 100
	if filled > width {
		filled = width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return rego.HStack(
		rego.Text(fmt.Sprintf("%-6s", label)).Bold(),
		rego.Text(bar).Color(color),
		rego.Text(fmt.Sprintf(" %3d%%", percent)),
	)
}

// =============================================================================
// 主应用
// =============================================================================

func App(c rego.C) rego.Node {
	// 获取模拟数据
	cpu, mem, net := useSystemStats(c)
	now := rego.Use(c, "now", time.Now())

	// 更新时间
	useInterval(c, func() {
		now.Set(time.Now())
	}, time.Second)

	// 日志列表
	logs := rego.Use(c, "logs", []string{
		"System boot successful.",
		"Initializing Rego engine...",
		"Hooks subsystem active.",
	})

	// 模拟日志增长
	useInterval(c, func() {
		newLogs := append(logs.Val, fmt.Sprintf("[%s] Event detected: %d", time.Now().Format("15:04:05"), rand.Intn(1000)))
		if len(newLogs) > 8 {
			newLogs = newLogs[1:]
		}
		logs.Set(newLogs)
	}, 2*time.Second)

	return rego.VStack(
		// 顶部：标题和时间
		rego.Box(
			rego.HStack(
				rego.Text("🚀 REGO SYSTEM MONITOR").Bold().Color(rego.Cyan),
				rego.Spacer(),
				rego.Text(now.Val.Format("2006-01-02 15:04:05")).Color(rego.Yellow),
			),
		).Border(rego.BorderDouble).BorderColor(rego.Cyan).Padding(0, 1),

		rego.Text(""),

		// 中间：状态和日志
		rego.HStack(
			// 左侧：实时状态
			rego.Box(
				rego.ScrollBox(c.Child("stats-scroll"),
					rego.VStack(
						rego.Text("📊 REAL-TIME STATS").Bold().Underline(),
						rego.Text(""),
						ProgressBar("CPU", cpu, rego.Red),
						rego.Text(""),
						ProgressBar("MEM", mem, rego.Green),
						rego.Text(""),
						ProgressBar("NET", net, rego.Blue),
						rego.Spacer(),
						rego.Text("Status: ONLINE").Color(rego.Green).Dim(),
					),
				),
			).Border(rego.BorderSingle).Padding(1, 2).Flex(1),

			rego.Text("  "),

			// 右侧：系统日志
			rego.Box(
				rego.ScrollBox(c.Child("logs-scroll"),
					rego.VStack(
						rego.Text("📜 SYSTEM LOGS").Bold().Underline(),
						rego.Text(""),
						rego.For(logs.Val, func(log string, i int) rego.Node {
							return rego.Text("> " + log).Dim()
						}),
						rego.Spacer(),
					),
				),
			).Border(rego.BorderSingle).Padding(1, 2).Flex(1),
		).Flex(1),

		rego.Text(""),

		// 底部：快捷键提示
		rego.Box(
			rego.HStack(
				rego.Text("Shortcuts: ").Dim(),
				rego.Text("[Ctrl+C] Quit").Bold(),
				rego.Spacer(),
				rego.Text("Built with Rego Hooks Engine").Italic().Dim(),
			),
		).Border(rego.BorderSingle).Padding(0, 1),
	)
}

func main() {
	if err := rego.Run(App); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
