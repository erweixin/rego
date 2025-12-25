package rego

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
)

// LargeApp 模拟一个拥有 1000 个节点的极其复杂的 TUI 界面
// 用于压力测试布局计算和渲染性能
func LargeApp(c C) Node {
	var rows []Node
	for i := 0; i < 50; i++ {
		var cols []Node
		for j := 0; j < 20; j++ {
			// 混合使用各种节点和布局属性
			cols = append(cols,
				Box(
					VStack(
						Text(fmt.Sprintf("R%d C%d", i, j)).Bold(),
						Divider().Color(Gray),
						Text("Data Point").Dim(),
					),
				).
					Border(BorderSingle).
					Width(12).
					Padding(0, 1),
			)
		}
		rows = append(rows, HStack(cols...).Gap(1))
	}
	return VStack(rows...).Gap(1)
}

// BenchmarkFullRender 测试从状态变化到屏幕最终绘制的完整链路
func BenchmarkFullRender(b *testing.B) {
	// 使用 SimulationScreen 避免真实的终端 I/O
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		b.Fatal(err)
	}
	screen.SetSize(200, 60)

	runtime := NewTestRuntime(LargeApp, screen)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.Render()
	}
}

// BenchmarkLayoutOnly 专门测试布局引擎计算高度的性能
func BenchmarkLayoutOnly(b *testing.B) {
	// 构造一个静态的巨大 Node 树
	root := LargeApp(&componentContext{
		children: make(map[string]*componentContext),
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 模拟 200 宽度的容器布局计算
		measureNodeHeight(root, 200)
	}
}

// BenchmarkTextWrapping 测试自动换行算法的性能
func BenchmarkTextWrapping(b *testing.B) {
	longText := "Rego 是一个 Hooks 风格的 Go CLI/TUI 框架，让你用类似 React 的方式构建终端应用。"
	for i := 0; i < 5; i++ {
		longText += longText // 构造极长文本
	}
	node := Text(longText).Wrap(true)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		measureNodeHeight(node, 40)
	}
}
