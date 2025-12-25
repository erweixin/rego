package rego

import "github.com/gdamore/tcell/v2"

// Color 表示颜色
type Color int

// 基础颜色常量
const (
	Default Color = iota
	Black
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	Gray
)

// Style 表示样式
type Style struct {
	fg        Color
	bg        Color
	bold      bool
	italic    bool
	underline bool
	dim       bool
	blink     bool

	// 布局属性
	width  int
	height int
	flex   int

	paddingTop    int
	paddingBottom int
	paddingLeft   int
	paddingRight  int

	border      BorderStyle
	borderColor Color

	align  Align
	valign Align
}

func defaultStyle() Style {
	return Style{
		fg:          Default,
		bg:          Default,
		borderColor: Default,
		align:       AlignLeft,
		valign:      AlignLeft,
	}
}

// NewStyle 创建一个新的样式对象
func NewStyle() Style {
	return defaultStyle()
}

// 链式方法

func (s Style) Foreground(c Color) Style {
	s.fg = c
	return s
}

func (s Style) Background(c Color) Style {
	s.bg = c
	return s
}

func (s Style) Bold() Style {
	s.bold = true
	return s
}

func (s Style) Italic() Style {
	s.italic = true
	return s
}

func (s Style) Underline() Style {
	s.underline = true
	return s
}

func (s Style) Dim() Style {
	s.dim = true
	return s
}

func (s Style) Blink() Style {
	s.blink = true
	return s
}

func (s Style) Width(w int) Style {
	s.width = w
	return s
}

func (s Style) Height(h int) Style {
	s.height = h
	return s
}

func (s Style) Flex(f int) Style {
	s.flex = f
	return s
}

func (s Style) Padding(v, h int) Style {
	s.paddingTop = v
	s.paddingBottom = v
	s.paddingLeft = h
	s.paddingRight = h
	return s
}

func (s Style) PaddingAll(top, right, bottom, left int) Style {
	s.paddingTop = top
	s.paddingRight = right
	s.paddingBottom = bottom
	s.paddingLeft = left
	return s
}

func (s Style) Border(style BorderStyle) Style {
	s.border = style
	return s
}

func (s Style) BorderColor(c Color) Style {
	s.borderColor = c
	return s
}

func (s Style) Align(a Align) Style {
	s.align = a
	return s
}

func (s Style) Valign(a Align) Style {
	s.valign = a
	return s
}

func (s Style) toTcell() tcell.Style {
	style := tcell.StyleDefault

	// 前景色
	style = style.Foreground(colorToTcell(s.fg))

	// 背景色
	if s.bg != Default {
		style = style.Background(colorToTcell(s.bg))
	}

	// 文字样式
	if s.bold {
		style = style.Bold(true)
	}
	if s.italic {
		style = style.Italic(true)
	}
	if s.underline {
		style = style.Underline(true)
	}
	if s.dim {
		style = style.Dim(true)
	}
	if s.blink {
		style = style.Blink(true)
	}

	return style
}

func colorToTcell(c Color) tcell.Color {
	switch c {
	case Black:
		return tcell.ColorBlack
	case Red:
		return tcell.ColorRed
	case Green:
		return tcell.ColorGreen
	case Yellow:
		return tcell.ColorYellow
	case Blue:
		return tcell.ColorBlue
	case Magenta:
		return tcell.ColorDarkMagenta
	case Cyan:
		return tcell.ColorDarkCyan
	case White:
		return tcell.ColorWhite
	case Gray:
		return tcell.ColorGray
	default:
		return tcell.ColorDefault
	}
}

// =============================================================================
// Border 样式
// =============================================================================

// BorderStyle 边框样式
type BorderStyle int

const (
	BorderNone BorderStyle = iota
	BorderSingle
	BorderDouble
	BorderRounded
	BorderThick
)

// BorderChars 边框字符
type BorderChars struct {
	TopLeft     rune
	TopRight    rune
	BottomLeft  rune
	BottomRight rune
	Horizontal  rune
	Vertical    rune
}

func getBorderChars(style BorderStyle) BorderChars {
	switch style {
	case BorderSingle:
		return BorderChars{
			TopLeft:     '┌',
			TopRight:    '┐',
			BottomLeft:  '└',
			BottomRight: '┘',
			Horizontal:  '─',
			Vertical:    '│',
		}
	case BorderDouble:
		return BorderChars{
			TopLeft:     '╔',
			TopRight:    '╗',
			BottomLeft:  '╚',
			BottomRight: '╝',
			Horizontal:  '═',
			Vertical:    '║',
		}
	case BorderRounded:
		return BorderChars{
			TopLeft:     '╭',
			TopRight:    '╮',
			BottomLeft:  '╰',
			BottomRight: '╯',
			Horizontal:  '─',
			Vertical:    '│',
		}
	case BorderThick:
		return BorderChars{
			TopLeft:     '┏',
			TopRight:    '┓',
			BottomLeft:  '┗',
			BottomRight: '┛',
			Horizontal:  '━',
			Vertical:    '┃',
		}
	default:
		return BorderChars{}
	}
}

// =============================================================================
// Align 对齐
// =============================================================================

// Align 对齐方式
type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)
