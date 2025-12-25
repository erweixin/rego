package rego

import (
	"github.com/gdamore/tcell/v2"
)

// =============================================================================
// Box 节点
// =============================================================================

type boxNode struct {
	child Node
	style Style
}

// Box 创建一个容器节点
func Box(child Node) *boxNode {
	return &boxNode{
		child: child,
		style: defaultStyle(),
	}
}

// Apply 应用样式
func (b *boxNode) Apply(s Style) *boxNode {
	b.style = s
	return b
}

func (b *boxNode) render(screen tcell.Screen, x, y, width, height int) int {
	if height <= 0 || width <= 0 {
		return 0
	}

	// 计算实际使用的宽高
	actualWidth := width
	actualHeight := height
	if b.style.width > 0 && b.style.width < width {
		actualWidth = b.style.width
	}
	if b.style.height > 0 && b.style.height < height {
		actualHeight = b.style.height
	}

	// 计算边框占用的空间
	borderSize := 0
	if b.style.border != BorderNone {
		borderSize = 1
	}

	// 计算内容最大可用区域
	maxContentWidth := actualWidth - borderSize*2 - b.style.paddingLeft - b.style.paddingRight
	maxContentHeight := actualHeight - borderSize*2 - b.style.paddingTop - b.style.paddingBottom

	if maxContentWidth <= 0 || maxContentHeight <= 0 {
		// 即使没有内容空间，也要绘制背景和边框
		if b.style.bg != Default {
			b.drawBackground(screen, x, y, actualWidth, actualHeight)
		}
		if b.style.border != BorderNone {
			b.renderBorder(screen, x, y, actualWidth, actualHeight)
		}
		return actualHeight
	}

	// 测量子节点实际需要的大小
	childHeight := 0
	if b.child != nil {
		childHeight = measureNodeHeight(b.child, maxContentWidth)
	}

	// 计算垂直起始位置 (Valign)
	contentY := y + borderSize + b.style.paddingTop
	if b.style.height > 0 || height > childHeight {
		// 如果有固定高度，或者给定高度大于内容高度，则可以进行垂直对齐
		availableH := maxContentHeight
		if childHeight < availableH {
			switch b.style.valign {
			case AlignCenter:
				contentY += (availableH - childHeight) / 2
			case AlignRight: // AlignRight 在垂直语境下作为 Bottom
				contentY += (availableH - childHeight)
			}
		}
	}

	contentX := x + borderSize + b.style.paddingLeft

	// 绘制背景
	if b.style.bg != Default {
		b.drawBackground(screen, x, y, actualWidth, actualHeight)
	}

	// 绘制边框
	if b.style.border != BorderNone {
		b.renderBorder(screen, x, y, actualWidth, actualHeight)
	}

	// 渲染子节点
	usedHeight := 0
	if b.child != nil {
		usedHeight = b.child.render(screen, contentX, contentY, maxContentWidth, maxContentHeight)
	}

	// 返回实际使用的高度
	if b.style.height > 0 {
		return b.style.height
	}
	return usedHeight + borderSize*2 + b.style.paddingTop + b.style.paddingBottom
}

func (b *boxNode) drawBackground(screen tcell.Screen, x, y, width, height int) {
	bgStyle := tcell.StyleDefault.Background(colorToTcell(b.style.bg))
	for row := y; row < y+height; row++ {
		for col := x; col < x+width; col++ {
			screen.SetContent(col, row, ' ', nil, bgStyle)
		}
	}
}

func (b *boxNode) renderBorder(screen tcell.Screen, x, y, width, height int) {
	chars := getBorderChars(b.style.border)
	style := tcell.StyleDefault.Foreground(colorToTcell(b.style.borderColor))

	// 四个角
	screen.SetContent(x, y, chars.TopLeft, nil, style)
	screen.SetContent(x+width-1, y, chars.TopRight, nil, style)
	screen.SetContent(x, y+height-1, chars.BottomLeft, nil, style)
	screen.SetContent(x+width-1, y+height-1, chars.BottomRight, nil, style)

	// 水平边
	for col := x + 1; col < x+width-1; col++ {
		screen.SetContent(col, y, chars.Horizontal, nil, style)
		screen.SetContent(col, y+height-1, chars.Horizontal, nil, style)
	}

	// 垂直边
	for row := y + 1; row < y+height-1; row++ {
		screen.SetContent(x, row, chars.Vertical, nil, style)
		screen.SetContent(x+width-1, row, chars.Vertical, nil, style)
	}
}

// 链式方法

// Width 设置固定宽度
func (b *boxNode) Width(w int) *boxNode {
	b.style.width = w
	return b
}

// Height 设置固定高度
func (b *boxNode) Height(h int) *boxNode {
	b.style.height = h
	return b
}

// Flex 设置 flex 权重
func (b *boxNode) Flex(f int) *boxNode {
	b.style.flex = f
	return b
}

// Padding 设置内边距 (垂直, 水平)
func (b *boxNode) Padding(vertical, horizontal int) *boxNode {
	b.style.paddingTop = vertical
	b.style.paddingBottom = vertical
	b.style.paddingLeft = horizontal
	b.style.paddingRight = horizontal
	return b
}

// PaddingAll 设置所有方向的内边距
func (b *boxNode) PaddingAll(top, right, bottom, left int) *boxNode {
	b.style.paddingTop = top
	b.style.paddingRight = right
	b.style.paddingBottom = bottom
	b.style.paddingLeft = left
	return b
}

// Border 设置边框样式
func (b *boxNode) Border(style BorderStyle) *boxNode {
	b.style.border = style
	return b
}

// BorderColor 设置边框颜色
func (b *boxNode) BorderColor(c Color) *boxNode {
	b.style.borderColor = c
	return b
}

// Align 设置对齐方式
func (b *boxNode) Align(a Align) *boxNode {
	b.style.align = a
	return b
}

// Valign 设置垂直对齐方式
func (b *boxNode) Valign(a Align) *boxNode {
	b.style.valign = a
	return b
}

// Background 设置背景色
func (b *boxNode) Background(c Color) *boxNode {
	b.style.bg = c
	return b
}

// Width 设置文本固定宽度
func (t *textNode) Width(w int) *textNode {
	t.style.width = w
	return t
}

// Flex 设置文本 flex 权重
func (t *textNode) Flex(f int) *textNode {
	t.style.flex = f
	return t
}

// Align 设置文本对齐
func (t *textNode) Align(a Align) *textNode {
	t.style.align = a
	return t
}
