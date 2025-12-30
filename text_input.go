package rego

import (
	"strings"
	"unicode/utf8"

	"github.com/mattn/go-runewidth"
)

// =============================================================================
// TextInput - 文本输入组件 (增强支持多行)
// =============================================================================

type TextInputProps struct {
	Value       string
	Placeholder string
	Label       string
	Width       int
	Height      int  // 0 表示单行，>1 表示多行
	Multiline   bool // 是否开启多行模式
	OnChanged   func(string)
	OnSubmit    func(string)
	Password    bool // 是否为密码模式
}

func TextInput(c C, props TextInputProps) Node {
	focus := UseFocus(c)
	text := Use(c, "text", props.Value)
	// 在多行模式下，cursorPos 是整个字符串的 rune 偏移量
	cursorPos := Use(c, "cursorPos", utf8.RuneCountInString(text.Val))

	// 同步外部 Value
	UseEffect(c, func() func() {
		if props.Value != text.Val {
			text.Set(props.Value)
			// 注意：这里简单重置光标到末尾，或者保持不变
		}
		return nil
	}, props.Value)

	// 鼠标点击处理
	UseMouse(c, func(ev MouseEvent) {
		if ev.Type != MouseEventClick || ev.Button != MouseButtonLeft {
			return
		}

		rect := c.Rect()
		if !rect.Contains(ev.X, ev.Y) {
			return
		}

		// 点击时聚焦
		focus.Focus()

		// 计算文本区域的偏移量
		// 布局: Box > VStack > [Label?] > Box(border+padding) > content
		// border: 1, padding: (0, 1)
		textAreaX := rect.X + 1 + 1 // border + padding
		textAreaY := rect.Y + 1     // border
		if props.Label != "" {
			textAreaY += 1 // Label 占一行
		}

		// 计算点击的相对位置
		clickCol := ev.X - textAreaX
		clickRow := ev.Y - textAreaY

		if clickCol < 0 {
			clickCol = 0
		}
		if clickRow < 0 {
			clickRow = 0
		}

		// 根据点击位置计算光标位置
		displayVal := text.Val
		if props.Password {
			displayVal = strings.Repeat("*", utf8.RuneCountInString(text.Val))
		}

		newPos := calculateCursorPosFromClick(displayVal, clickRow, clickCol)
		cursorPos.Set(newPos)
	})

	// 键盘处理
	UseKey(c, func(key Key, r rune) {
		if !focus.IsFocused {
			return
		}

		runes := []rune(text.Val)
		currentLen := len(runes)

		switch key {
		case KeyBackspace:
			if cursorPos.Val > 0 {
				newRunes := append(runes[:cursorPos.Val-1], runes[cursorPos.Val:]...)
				newVal := string(newRunes)
				text.Set(newVal)
				cursorPos.Update(func(v int) int { return v - 1 })
				if props.OnChanged != nil {
					props.OnChanged(newVal)
				}
			}
		case KeyDelete:
			if cursorPos.Val < currentLen {
				newRunes := append(runes[:cursorPos.Val], runes[cursorPos.Val+1:]...)
				newVal := string(newRunes)
				text.Set(newVal)
				if props.OnChanged != nil {
					props.OnChanged(newVal)
				}
			}
		case KeyLeft:
			if cursorPos.Val > 0 {
				cursorPos.Update(func(v int) int { return v - 1 })
			}
		case KeyRight:
			if cursorPos.Val < currentLen {
				cursorPos.Update(func(v int) int { return v + 1 })
			}
		case KeyUp:
			if props.Multiline {
				// 找到上一行的位置
				cursorPos.Set(findPosAbove(runes, cursorPos.Val))
			}
		case KeyDown:
			if props.Multiline {
				// 找到下一行的位置
				cursorPos.Set(findPosBelow(runes, cursorPos.Val))
			}
		case KeyEnter:
			if props.Multiline {
				// 多行模式下 Enter 是换行
				newRunes := make([]rune, 0, len(runes)+1)
				newRunes = append(newRunes, runes[:cursorPos.Val]...)
				newRunes = append(newRunes, '\n')
				newRunes = append(newRunes, runes[cursorPos.Val:]...)
				newVal := string(newRunes)
				text.Set(newVal)
				cursorPos.Update(func(v int) int { return v + 1 })
				if props.OnChanged != nil {
					props.OnChanged(newVal)
				}
			} else {
				if props.OnSubmit != nil {
					props.OnSubmit(text.Val)
				}
			}
		case KeyHome:
			// 跳转到行首（多行模式跳转到当前行行首）
			if props.Multiline {
				lineStart := cursorPos.Val
				for lineStart > 0 && runes[lineStart-1] != '\n' {
					lineStart--
				}
				cursorPos.Set(lineStart)
			} else {
				cursorPos.Set(0)
			}
		case KeyEnd:
			// 跳转到行尾
			if props.Multiline {
				lineEnd := cursorPos.Val
				for lineEnd < len(runes) && runes[lineEnd] != '\n' {
					lineEnd++
				}
				cursorPos.Set(lineEnd)
			} else {
				cursorPos.Set(currentLen)
			}
		default:
			if r != 0 {
				newRunes := make([]rune, 0, len(runes)+1)
				newRunes = append(newRunes, runes[:cursorPos.Val]...)
				newRunes = append(newRunes, r)
				newRunes = append(newRunes, runes[cursorPos.Val:]...)
				newVal := string(newRunes)
				text.Set(newVal)
				cursorPos.Update(func(v int) int { return v + 1 })
				if props.OnChanged != nil {
					props.OnChanged(newVal)
				}
			}
		}
	})

	// 渲染逻辑
	displayVal := text.Val
	if props.Password {
		displayVal = strings.Repeat("*", utf8.RuneCountInString(text.Val))
	}

	runes := []rune(displayVal)
	before := string(runes[:cursorPos.Val])
	after := ""
	if cursorPos.Val < len(runes) {
		after = string(runes[cursorPos.Val:])
	}

	// 将文本按行分割渲染
	linesBefore := strings.Split(before, "\n")
	linesAfter := strings.Split(after, "\n")

	// 构造多行视图
	var rows []Node

	// 处理光标所在行之前的所有行
	for i := 0; i < len(linesBefore)-1; i++ {
		rows = append(rows, Text(linesBefore[i]))
	}

	// 处理光标所在行 (HStack 拼接光标前、硬件光标、光标后)
	currentLineBefore := linesBefore[len(linesBefore)-1]
	currentLineAfter := linesAfter[0]
	rows = append(rows, HStack(
		Text(currentLineBefore),
		When(focus.IsFocused, Cursor(c)),
		Text(currentLineAfter),
	))

	// 处理光标之后的所有行
	for i := 1; i < len(linesAfter); i++ {
		rows = append(rows, Text(linesAfter[i]))
	}

	var content Node = VStack(rows...)
	if text.Val == "" && !focus.IsFocused {
		placeholder := Text(props.Placeholder).Dim()
		if props.Multiline {
			content = VStack(placeholder)
		} else {
			content = placeholder
		}
	}

	// 计算容器高度
	boxHeight := props.Height
	if boxHeight == 0 {
		if props.Multiline {
			boxHeight = 6 // 多行默认高度
		} else {
			boxHeight = 3 // 单行默认高度
		}
	}

	return c.Wrap(Box(
		VStack(
			When(props.Label != "", Text(props.Label).Dim().Bold()),
			Box(WhenElse(props.Multiline, ScrollBox(c.Child("scroll"), content), content)).
				Padding(0, 1).
				Border(BorderSingle).
				BorderColor(If(focus.IsFocused, Cyan, Gray)).
				Height(boxHeight),
		),
	).Width(props.Width))
}

// 辅助函数：根据点击的行列计算光标位置
func calculateCursorPosFromClick(text string, clickRow, clickCol int) int {
	lines := strings.Split(text, "\n")

	// 限制行号范围
	if clickRow < 0 {
		clickRow = 0
	}
	if clickRow >= len(lines) {
		clickRow = len(lines) - 1
	}

	// 计算点击行之前所有字符的数量（包括换行符）
	pos := 0
	for i := 0; i < clickRow; i++ {
		pos += utf8.RuneCountInString(lines[i]) + 1 // +1 for '\n'
	}

	// 在当前行中根据显示宽度找到对应的字符位置
	line := lines[clickRow]
	lineRunes := []rune(line)
	currentWidth := 0

	for i, r := range lineRunes {
		charWidth := runewidth.RuneWidth(r)
		// 如果点击位置在当前字符的范围内
		if currentWidth+charWidth > clickCol {
			pos += i
			return pos
		}
		currentWidth += charWidth
	}

	// 点击在行尾之后，光标放在行尾
	pos += len(lineRunes)
	return pos
}

// 辅助函数：计算上一行对应位置
func findPosAbove(runes []rune, current int) int {
	if current == 0 {
		return 0
	}
	// 1. 找到当前行行首
	lineStart := current
	for lineStart > 0 && runes[lineStart-1] != '\n' {
		lineStart--
	}
	if lineStart == 0 {
		return 0
	} // 已经是第一行

	// 2. 找到上一行行首
	prevLineEnd := lineStart - 1
	prevLineStart := prevLineEnd
	for prevLineStart > 0 && runes[prevLineStart-1] != '\n' {
		prevLineStart--
	}

	// 3. 尽量保持横坐标一致
	col := current - lineStart
	prevLineLen := prevLineEnd - prevLineStart
	if col > prevLineLen {
		return prevLineEnd
	}
	return prevLineStart + col
}

// 辅助函数：计算下一行对应位置
func findPosBelow(runes []rune, current int) int {
	// 1. 找到当前行行首
	lineStart := current
	for lineStart > 0 && runes[lineStart-1] != '\n' {
		lineStart--
	}
	// 2. 找到当前行行尾
	lineEnd := current
	for lineEnd < len(runes) && runes[lineEnd] != '\n' {
		lineEnd++
	}
	if lineEnd == len(runes) {
		return len(runes)
	} // 已经是最后一行

	// 3. 找到下一行结束位置
	nextLineStart := lineEnd + 1
	nextLineEnd := nextLineStart
	for nextLineEnd < len(runes) && runes[nextLineEnd] != '\n' {
		nextLineEnd++
	}

	// 4. 尽量保持横坐标一致
	col := current - lineStart
	nextLineLen := nextLineEnd - nextLineStart
	if col > nextLineLen {
		return nextLineEnd
	}
	return nextLineStart + col
}
