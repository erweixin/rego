package rego

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// drawAnsiString renders a string containing ANSI escape codes to tcell.Screen
// It handles a subset of ANSI codes commonly used for TUI styling.
func drawAnsiString(screen tcell.Screen, x, y, width, height int, text string, baseStyle tcell.Style) int {
	return renderAnsi(screen, x, y, width, height, text, baseStyle)
}

func renderAnsi(screen tcell.Screen, x, y, width, height int, text string, baseStyle tcell.Style) int {
	currentStyle := baseStyle
	curX, curY := x, y
	lines := 1

	// 基于 rune 的处理，对普通文字使用 runewidth
	runes := []rune(text)
	for i := 0; i < len(runes); i++ {
		r := runes[i]

		if r == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			j := i + 2
			for j < len(runes) && (runes[j] < 0x40 || runes[j] > 0x7E) {
				j++
			}
			if j < len(runes) && runes[j] == 'm' {
				params := runes[i+2 : j]
				currentStyle = applySGR(currentStyle, params)
				i = j
				continue
			}
			i = j
			continue
		}

		if r == '\n' {
			curX = x
			curY++
			lines++
			if lines > height {
				break
			}
			continue
		}

		w := runewidth.RuneWidth(r)
		if curX+w > x+width {
			curX = x
			curY++
			lines++
			if lines > height {
				break
			}
		}

		if curY < y+height {
			screen.SetContent(curX, curY, r, nil, currentStyle)
		}
		curX += w
	}

	return lines
}

func applySGR(style tcell.Style, params []rune) tcell.Style {
	if len(params) == 0 {
		return tcell.StyleDefault
	}

	parts := stringsSplitRunes(params, ';')
	for i := 0; i < len(parts); i++ {
		p := parts[i]
		code := 0
		if len(p) > 0 {
			code = atoi(string(p))
		}

		switch code {
		case 0: // Reset
			style = tcell.StyleDefault
		case 1: // Bold
			style = style.Bold(true)
		case 2: // Dim
			style = style.Dim(true)
		case 3: // Italic
			style = style.Italic(true)
		case 4: // Underline
			style = style.Underline(true)
		case 5, 6: // Blink
			style = style.Blink(true)
		case 7: // Reverse
			style = style.Reverse(true)
		case 22: // Normal intensity
			style = style.Bold(false).Dim(false)
		case 23: // Italic off
			style = style.Italic(false)
		case 24: // Underline off
			style = style.Underline(false)
		case 30, 31, 32, 33, 34, 35, 36, 37: // Foreground
			style = style.Foreground(ansiColor(code - 30))
		case 38, 48: // Extended foreground/background
			// 38;5;n (256 colors) or 38;2;r;g;b (TrueColor)
			if i+2 < len(parts) {
				mode := atoi(string(parts[i+1]))
				if mode == 5 { // 256 colors
					colorCode := atoi(string(parts[i+2]))
					if code == 38 {
						style = style.Foreground(tcell.PaletteColor(colorCode))
					} else {
						style = style.Background(tcell.PaletteColor(colorCode))
					}
					i += 2
				} else if mode == 2 && i+4 < len(parts) { // TrueColor
					r := int32(atoi(string(parts[i+2])))
					g := int32(atoi(string(parts[i+3])))
					b := int32(atoi(string(parts[i+4])))
					color := tcell.NewRGBColor(r, g, b)
					if code == 38 {
						style = style.Foreground(color)
					} else {
						style = style.Background(color)
					}
					i += 4
				}
			}
		case 39: // Default foreground
			style = style.Foreground(tcell.ColorDefault)
		case 40, 41, 42, 43, 44, 45, 46, 47: // Background
			style = style.Background(ansiColor(code - 40))
		case 49: // Default background
			style = style.Background(tcell.ColorDefault)
		case 90, 91, 92, 93, 94, 95, 96, 97: // High intensity foreground
			style = style.Foreground(ansiColor(code - 90 + 8))
		case 100, 101, 102, 103, 104, 105, 106, 107: // High intensity background
			style = style.Background(ansiColor(code - 100 + 8))
		}
	}
	return style
}

func ansiColor(code int) tcell.Color {
	if code >= 0 && code < 16 {
		return tcell.PaletteColor(code)
	}
	return tcell.ColorDefault
}

func stringsSplitRunes(s []rune, sep rune) [][]rune {
	var res [][]rune
	start := 0
	for i, r := range s {
		if r == sep {
			res = append(res, s[start:i])
			start = i + 1
		}
	}
	res = append(res, s[start:])
	return res
}

func atoi(s string) int {
	res := 0
	for _, r := range s {
		if r >= '0' && r <= '9' {
			res = res*10 + int(r-'0')
		}
	}
	return res
}
