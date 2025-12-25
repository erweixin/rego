package rego

// =============================================================================
// Button - 按钮组件
// =============================================================================

type ButtonProps struct {
	Label   string
	OnClick func()
	Primary bool
}

func Button(c C, props ButtonProps) Node {
	focus := UseFocus(c)

	UseKey(c, func(key Key, r rune) {
		if focus.IsFocused && (key == KeyEnter || r == ' ') {
			if props.OnClick != nil {
				props.OnClick()
			}
		}
	})

	// 鼠标点击支持
	UseMouse(c, func(ev MouseEvent) {
		if ev.Type == MouseEventClick && ev.Button == MouseButtonLeft {
			if c.Rect().Contains(ev.X, ev.Y) {
				focus.Focus() // 点击聚焦
				if props.OnClick != nil {
					props.OnClick()
				}
			}
		}
	})

	label := "[" + props.Label + "]"
	style := Text(label)

	if props.Primary {
		style = style.Bold().Color(Cyan)
	}

	if focus.IsFocused {
		style = style.Background(Green).Color(Black)
	}

	return c.Wrap(Box(style).Padding(0, 1))
}
