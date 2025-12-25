package rego

import (
	"github.com/charmbracelet/glamour"
	"github.com/gdamore/tcell/v2"
)

// =============================================================================
// Markdown 节点
// =============================================================================

type markdownNode struct {
	content string
	style   Style
	theme   string // glamour theme: "dark", "light", "notty", etc.

	// 缓存机制，避免频繁调用 glamour 导致卡顿
	lastContent string
	lastWidth   int
	lastHeight  int
	lastOutput  string
}

// Markdown 创建一个 Markdown 渲染节点
func Markdown(content string) *markdownNode {
	return &markdownNode{
		content: content,
		style:   defaultStyle(),
		theme:   "dark",
	}
}

// Theme 设置 glamour 主题
func (m *markdownNode) Theme(theme string) *markdownNode {
	m.theme = theme
	return m
}

// Apply 应用样式
func (m *markdownNode) Apply(s Style) *markdownNode {
	m.style = s
	return m
}

func (m *markdownNode) render(screen tcell.Screen, x, y, width, height int) int {
	if width <= 0 || height <= 0 {
		return 0
	}

	out := m.getRenderedOutput(width)
	return renderAnsi(screen, x, y, width, height, out, m.style.toTcell())
}

func (m *markdownNode) getRenderedOutput(width int) string {
	// 只有内容、宽度都一致时才使用缓存
	if m.lastOutput != "" && m.lastWidth == width && m.lastContent == m.content {
		return m.lastOutput
	}

	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(m.theme),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return m.content
	}

	out, err := r.Render(m.content)
	if err != nil {
		return m.content
	}

	m.lastOutput = out
	m.lastWidth = width
	m.lastContent = m.content

	// 计算并缓存高度
	lines := 0
	for _, r := range out {
		if r == '\n' {
			lines++
		}
	}
	m.lastHeight = lines + 1

	return out
}

// 实现 flexNode 接口
func (m *markdownNode) getFlex() int {
	return m.style.flex
}

func (m *markdownNode) getHeight() int {
	if m.style.height > 0 {
		return m.style.height
	}
	if m.lastHeight > 0 {
		return m.lastHeight
	}
	return 0
}

// measureHeight 辅助函数用于测量高度
func (m *markdownNode) measureHeight(width int) int {
	m.getRenderedOutput(width)
	return m.lastHeight
}

// measureMarkdownHeight 保持兼容性
func measureMarkdownHeight(content string, width int, theme string) int {
	m := Markdown(content).Theme(theme)
	return m.measureHeight(width)
}
