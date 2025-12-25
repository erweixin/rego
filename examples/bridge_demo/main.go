package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/erweixin/rego"
	"github.com/erweixin/rego/examples/bridge_demo/core"
)

// 1. 定义全局 Context
type AppBridge = *rego.Bridge[core.AppState, core.Question, bool]

var BridgeContext = rego.CreateContext[AppBridge](nil)

// -----------------------------------------------------------------------------
// 子组件示例：深层嵌套组件
// -----------------------------------------------------------------------------

// 最深层的交互组件
func InteractionArea(c rego.C) rego.Node {
	// 通过 Context 获取 Bridge，不需要从 Props 传进来
	bridge := rego.UseContext(c, BridgeContext)
	if bridge == nil || !bridge.HasInteraction() {
		return rego.Text(" [ 状态：运行中... ] ").Dim()
	}

	q := bridge.Interaction()
	return rego.Box(
		rego.VStack(
			rego.Text(" ⚠️  来自深层组件的确认请求:").Bold().Color(rego.Yellow),
			rego.Text(q.Body).Italic(),
			rego.Text(""),
			rego.HStack(
				rego.Button(c.Child("yes"), rego.ButtonProps{
					Label:   "确认",
					OnClick: func() { bridge.Submit(true) },
					Primary: true,
				}),
				rego.Text("  "),
				rego.Button(c.Child("no"), rego.ButtonProps{
					Label:   "拒绝",
					OnClick: func() { bridge.Submit(false) },
				}),
			).Gap(2),
		),
	).Border(rego.BorderRounded).Padding(1, 1).BorderColor(rego.Yellow)
}

// 中间层：工作区
func Workspace(c rego.C) rego.Node {
	bridge := rego.UseContext(c, BridgeContext)
	if bridge == nil {
		return rego.Text("Workspace Error").Color(rego.Red)
	}
	state := bridge.State()

	return rego.VStack(
		rego.Text(" >>> 当前工作区 <<< ").Dim(),
		rego.Box(
			rego.VStack(
				rego.Text("日志记录:").Bold(),
				rego.ScrollBox(c.Child("logs"),
					rego.For(state.Logs, func(log string, i int) rego.Node {
						return rego.Text(" • " + log)
					}),
				).Flex(1),
			),
		).Border(rego.BorderSingle).Flex(1).Padding(0, 1),
		rego.Text(""),
		// 渲染深层交互组件
		InteractionArea(c.Child("interaction")),
	).Flex(1)
}

// 侧边栏：显示状态信息
func Sidebar(c rego.C) rego.Node {
	bridge := rego.UseContext(c, BridgeContext)
	if bridge == nil {
		return rego.Text("Sidebar Error").Color(rego.Red)
	}
	state := bridge.State()

	return rego.Box(
		rego.VStack(
			rego.Text("「侧栏信息」").Bold().Color(rego.Cyan),
			rego.Divider(),
			rego.Text("状态: "+state.Status),
			rego.Text(fmt.Sprintf("进度: %d%%", state.Progress)),
			rego.Spacer(),
			rego.Text("V1.0.0").Dim(),
		),
	).Width(20).Border(rego.BorderSingle).Padding(0, 1)
}

// -----------------------------------------------------------------------------
// UI 适配器：根组件
// -----------------------------------------------------------------------------

func App(c rego.C) rego.Node {
	// 创建 Bridge
	bridge := rego.UseBridge[core.AppState, core.Question, bool](c, core.AppState{Status: "等待启动"})

	// 启动 Core
	rego.UseEffect(c, func() func() {
		go core.Run(bridge.Handle())
		return nil
	}, []any{})

	// 使用 BridgeContext.Provide 将 bridge 注入到整个组件树
	return BridgeContext.Provide(c, bridge,
		rego.Box(
			rego.VStack(
				// 顶部状态栏
				rego.HStack(
					rego.Text(" 🛠️  AGENT BRIDGE DASHBOARD ").Bold().Background(rego.Blue).Color(rego.White),
					rego.Spacer(),
					rego.When(!bridge.HasInteraction() && bridge.State().Status == "清理完成",
						rego.Button(c.Child("quit"), rego.ButtonProps{
							Label:   "退出程序",
							OnClick: func() { c.Quit() },
							Primary: true,
						}),
					),
				).Padding(0, 1),

				rego.Text(""),

				// 主体布局：左侧边栏 + 右侧工作区
				rego.HStack(
					Sidebar(c.Child("sidebar")),
					rego.Text(" "),
					Workspace(c.Child("workspace")),
				).Flex(1),
			),
		).Padding(1, 2),
	)
}

// -----------------------------------------------------------------------------
// CLI 适配器
// -----------------------------------------------------------------------------

type CLIHandler struct{}

func (h *CLIHandler) Update(s core.AppState) {
	fmt.Printf("\r[%-30s] %s | 进度: %d%%", strings.Repeat("#", s.Progress/4), s.Status, s.Progress)
}

func (h *CLIHandler) Ask(q core.Question) bool {
	fmt.Printf("\n\n>>> %s <<<\n%s (y/n): ", q.Title, q.Body)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}

func main() {
	isCLI := flag.Bool("cli", false, "以命令行模式运行")
	flag.Parse()

	if *isCLI {
		core.Run(&CLIHandler{})
	} else {
		rego.Run(App)
	}
}
