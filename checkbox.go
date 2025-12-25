package rego

// =============================================================================
// Checkbox - 复选框组件
// =============================================================================

type CheckboxProps struct {
	Label     string
	Checked   bool
	OnChanged func(bool)
}

func Checkbox(c C, props CheckboxProps) Node {
	focus := UseFocus(c)

	UseKey(c, func(key Key, r rune) {
		if focus.IsFocused && (key == KeyEnter || r == ' ') {
			if props.OnChanged != nil {
				props.OnChanged(!props.Checked)
			}
		}
	})

	// 鼠标点击支持
	UseMouse(c, func(ev MouseEvent) {
		if ev.Type == MouseEventClick && ev.Button == MouseButtonLeft {
			if c.Rect().Contains(ev.X, ev.Y) {
				focus.Focus() // 点击聚焦
				if props.OnChanged != nil {
					props.OnChanged(!props.Checked)
				}
			}
		}
	})

	icon := "[ ]"
	if props.Checked {
		icon = "[x]"
	}

	style := Text(icon + " " + props.Label)
	if focus.IsFocused {
		style = style.Color(Green)
	}

	return c.Wrap(Box(style).Padding(0, 1))
}
