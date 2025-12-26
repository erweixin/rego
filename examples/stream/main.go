package main

import (
	"fmt"
	"time"

	rego "github.com/erweixin/rego"
)

// =============================================================================
// StreamDemo - A demo purely showcasing streaming output and smart scrolling
// =============================================================================

func main() {
	if err := rego.Run(App); err != nil {
		panic(err)
	}
}

func App(c rego.C) rego.Node {
	messages := rego.Use(c, "messages", []string{})
	currentStream := rego.Use(c, "currentStream", "")
	isStreaming := rego.Use(c, "isStreaming", false)

	// Auto-start the first streaming task
	rego.UseEffect(c, func() func() {
		if !isStreaming.Val && len(messages.Val) == 0 {
			startDemoStream(c, messages, currentStream, isStreaming)
		}
		return nil
	}, isStreaming.Val)

	rego.UseKey(c, func(key rego.Key, r rune) {
		if r == 'r' && !isStreaming.Val {
			// Press R to reset and restart
			messages.Set([]string{})
			currentStream.Set("")
			startDemoStream(c, messages, currentStream, isStreaming)
		}
		if key == rego.KeyCtrlC {
			c.Quit()
		}
	})

	return rego.VStack(
		// Top title bar
		rego.Box(
			rego.HStack(
				rego.Text("🚀 REGO STREAMING DEMO").Bold().Color(rego.Cyan),
				rego.Spacer(),
				// rego.Stats(c.Child("stats")),
				rego.Text(" "),
				rego.Text(fmt.Sprintf("Status: %s", If(isStreaming.Val, "Streaming...", "Ready"))).
					Color(If(isStreaming.Val, rego.Yellow, rego.Green)),
			),
		).Border(rego.BorderSingle).Padding(0, 1),

		rego.Text(""),

		// Main view area: showcases the core logic of StreamView
		rego.TailBox(c.Child("chat-scroll"),
			rego.Box(
				rego.VStack(
					// Finished message history
					rego.For(messages.Val, func(msg string, i int) rego.Node {
						return rego.VStack(
							rego.Text(fmt.Sprintf("--- History Message #%d ---", i+1)).Dim(),
							rego.Markdown(msg),
							rego.Text(""),
						)
					}),

					// Current streaming message
					rego.When(isStreaming.Val || currentStream.Val != "",
						rego.VStack(
							rego.Text("--- AI is typing ---").Color(rego.Yellow).Italic(),
							rego.Markdown(currentStream.Val+"▍"),
						),
					),
				),
			).Apply(rego.NewStyle().Padding(1, 2)),
		).Flex(1),

		rego.Text(""),

		// Bottom operation hints
		rego.HStack(
			rego.Text(" [R] Restart Demo ").Background(rego.Blue).Color(rego.White),
			rego.Text("  "),
			rego.Text(" Ctrl+C Quit ").Color(rego.White),
			rego.Spacer(),
			rego.Text("Tip: Try scrolling up during generation").Dim(),
		),
	).Padding(1, 2)
}

// startDemoStream starts a simulated long text stream
func startDemoStream(c rego.C, history *rego.State[[]string], current *rego.State[string], status *rego.State[bool]) {
	status.Set(true)
	current.Set("")

	go func() {
		content := `
## Demonstrating Smart Scrolling (Auto-Tail)

When content grows via streaming, ` + "`TailBox`" + ` ensures your viewport always follows the latest tokens.

### Why is this important?
1. **No manual scrolling**: Agent output is fast, manual scrolling is exhausting.
2. **No flickering**: Even if the Markdown height changes constantly, Rego stays stable.

### Complex Content Test
Below is a highlighted code block. Observe the viewport behavior as it grows:

` + "```go" + `
package main

import "fmt"

func demonstrate() {
    for i := 0; i < 10; i++ {
        fmt.Printf("Token sequence: %d\n", i)
        // This code block will keep growing
    }
}
` + "```" + `

### List Growth
- Auto-generated item 1
- Auto-generated item 2
- Auto-generated item 3
- Auto-generated item 4
- Auto-generated item 5
- Auto-generated item 6
- Auto-generated item 7
- Auto-generated item 8
- Auto-generated item 9
- Auto-generated item 10
- Auto-generated item 1
- Auto-generated item 2
- Auto-generated item 3
- Auto-generated item 4
- Auto-generated item 5
- Auto-generated item 6
- Auto-generated item 7
- Auto-generated item 8
- Auto-generated item 9
- Auto-generated item 10
- Auto-generated item 1
- Auto-generated item 2
- Auto-generated item 3
- Auto-generated item 4
- Auto-generated item 5
- Auto-generated item 6
- Auto-generated item 7
- Auto-generated item 8
- Auto-generated item 9
- Auto-generated item 10

### Summary
This is the logic behind Rego's ` + "`StreamView`" + ` component.
It brings the Agent Developer Experience (DX) to an industrial grade.
`
		// Simulate tokens streaming out one by one
		fullRunes := []rune(content)
		currentText := ""
		for _, r := range fullRunes {
			currentText += string(r)
			current.Set(currentText)
			// Simulate random token generation speed
			time.Sleep(20 * time.Millisecond)
		}

		// Complete stream, save to history
		time.Sleep(500 * time.Millisecond)
		history.Update(func(h []string) []string {
			return append(h, currentText)
		})
		current.Set("")
		status.Set(false)
		c.Refresh()
	}()
}

// Simple helper function
func If[T any](cond bool, t, f T) T {
	if cond {
		return t
	}
	return f
}
