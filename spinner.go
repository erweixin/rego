package rego

import (
	"time"
)

// =============================================================================
// Spinner - 加载动画组件
// =============================================================================

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func Spinner(c C, label string) Node {
	frame := Use(c, "frame", 0)

	UseEffect(c, func() func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		go func() {
			for range ticker.C {
				frame.Update(func(v int) int {
					return (v + 1) % len(spinnerFrames)
				})
				c.Refresh()
			}
		}()
		return func() { ticker.Stop() }
	})

	return HStack(
		Text(spinnerFrames[frame.Val]).Color(Cyan),
		Text(" "+label),
	)
}
