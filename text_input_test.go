package rego

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

// =============================================================================
// 辅助函数测试
// =============================================================================

func TestCalculateCursorPosFromClick(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		clickRow int
		clickCol int
		expected int
	}{
		// 单行测试
		{
			name:     "单行-点击开头",
			text:     "hello",
			clickRow: 0,
			clickCol: 0,
			expected: 0,
		},
		{
			name:     "单行-点击中间",
			text:     "hello",
			clickRow: 0,
			clickCol: 2,
			expected: 2,
		},
		{
			name:     "单行-点击末尾",
			text:     "hello",
			clickRow: 0,
			clickCol: 5,
			expected: 5,
		},
		{
			name:     "单行-点击超出末尾",
			text:     "hello",
			clickRow: 0,
			clickCol: 10,
			expected: 5,
		},
		// 中文测试（每个中文占2列宽度）
		{
			name:     "中文-点击第一个字符",
			text:     "你好世界",
			clickRow: 0,
			clickCol: 0,
			expected: 0,
		},
		{
			name:     "中文-点击第一个字符中间",
			text:     "你好世界",
			clickRow: 0,
			clickCol: 1,
			expected: 0,
		},
		{
			name:     "中文-点击第二个字符开头",
			text:     "你好世界",
			clickRow: 0,
			clickCol: 2,
			expected: 1,
		},
		{
			name:     "中文-点击末尾",
			text:     "你好世界",
			clickRow: 0,
			clickCol: 8,
			expected: 4,
		},
		// 多行测试
		{
			name:     "多行-第一行开头",
			text:     "line1\nline2\nline3",
			clickRow: 0,
			clickCol: 0,
			expected: 0,
		},
		{
			name:     "多行-第二行开头",
			text:     "line1\nline2\nline3",
			clickRow: 1,
			clickCol: 0,
			expected: 6, // "line1\n" = 6 chars
		},
		{
			name:     "多行-第二行中间",
			text:     "line1\nline2\nline3",
			clickRow: 1,
			clickCol: 3,
			expected: 9, // "line1\n" + "lin" = 6 + 3
		},
		{
			name:     "多行-第三行",
			text:     "line1\nline2\nline3",
			clickRow: 2,
			clickCol: 2,
			expected: 14, // "line1\nline2\n" + "li" = 12 + 2
		},
		// 边界情况
		{
			name:     "空字符串",
			text:     "",
			clickRow: 0,
			clickCol: 0,
			expected: 0,
		},
		{
			name:     "负数行号",
			text:     "hello",
			clickRow: -1,
			clickCol: 0,
			expected: 0,
		},
		{
			name:     "超出行号",
			text:     "line1\nline2",
			clickRow: 5,
			clickCol: 0,
			expected: 6, // 应该定位到最后一行开头
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCursorPosFromClick(tt.text, tt.clickRow, tt.clickCol)
			if got != tt.expected {
				t.Errorf("calculateCursorPosFromClick(%q, %d, %d) = %d, want %d",
					tt.text, tt.clickRow, tt.clickCol, got, tt.expected)
			}
		})
	}
}

func TestFindPosAbove(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		current  int
		expected int
	}{
		{
			name:     "第一行-无法向上",
			text:     "hello",
			current:  2,
			expected: 0,
		},
		{
			name:     "第二行开头-移到第一行开头",
			text:     "line1\nline2",
			current:  6, // "line2" 开头
			expected: 0, // "line1" 开头
		},
		{
			name:     "第二行中间-保持列位置",
			text:     "line1\nline2",
			current:  8, // "li|ne2"
			expected: 2, // "li|ne1"
		},
		{
			name:     "第二行超过第一行长度",
			text:     "ab\nline2",
			current:  6, // "lin|e2"
			expected: 2, // "ab" 末尾
		},
		{
			name:     "三行-从第三行到第二行",
			text:     "line1\nline2\nline3",
			current:  14, // "li|ne3"
			expected: 8,  // "li|ne2"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.text)
			got := findPosAbove(runes, tt.current)
			if got != tt.expected {
				t.Errorf("findPosAbove(%q, %d) = %d, want %d",
					tt.text, tt.current, got, tt.expected)
			}
		})
	}
}

func TestFindPosBelow(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		current  int
		expected int
	}{
		{
			name:     "最后一行-无法向下",
			text:     "hello",
			current:  2,
			expected: 5, // 移到末尾
		},
		{
			name:     "第一行开头-移到第二行开头",
			text:     "line1\nline2",
			current:  0,
			expected: 6,
		},
		{
			name:     "第一行中间-保持列位置",
			text:     "line1\nline2",
			current:  2, // "li|ne1"
			expected: 8, // "li|ne2"
		},
		{
			name:     "第一行超过第二行长度",
			text:     "line1\nab",
			current:  4, // "line|1"
			expected: 8, // "ab" 末尾
		},
		{
			name:     "三行-从第一行到第二行",
			text:     "line1\nline2\nline3",
			current:  2, // "li|ne1"
			expected: 8, // "li|ne2"
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runes := []rune(tt.text)
			got := findPosBelow(runes, tt.current)
			if got != tt.expected {
				t.Errorf("findPosBelow(%q, %d) = %d, want %d",
					tt.text, tt.current, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// 混合中英文测试
// =============================================================================

func TestCalculateCursorPosFromClick_MixedContent(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		clickRow int
		clickCol int
		expected int
	}{
		{
			name:     "中英混合-点击英文部分",
			text:     "Hello你好",
			clickRow: 0,
			clickCol: 3, // "Hel" 之后
			expected: 3,
		},
		{
			name:     "中英混合-点击第一个中文",
			text:     "Hello你好",
			clickRow: 0,
			clickCol: 5, // "Hello" 之后，中文开始
			expected: 5,
		},
		{
			name:     "中英混合-点击第二个中文",
			text:     "Hello你好",
			clickRow: 0,
			clickCol: 7, // "Hello你" 之后 (5+2=7)
			expected: 6,
		},
		{
			name:     "多行中英混合",
			text:     "Hello\n你好世界",
			clickRow: 1,
			clickCol: 4, // 第二行，"你好" 之后 (2+2=4)
			expected: 8, // "Hello\n" + "你好" = 6 + 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateCursorPosFromClick(tt.text, tt.clickRow, tt.clickCol)
			if got != tt.expected {
				t.Errorf("calculateCursorPosFromClick(%q, %d, %d) = %d, want %d",
					tt.text, tt.clickRow, tt.clickCol, got, tt.expected)
			}
		})
	}
}

// =============================================================================
// Snapshot 测试辅助函数
// =============================================================================

func newTestScreen(w, h int) tcell.SimulationScreen {
	s := tcell.NewSimulationScreen("")
	s.Init()
	s.SetSize(w, h)
	return s
}

func getScreenContent(s tcell.SimulationScreen) string {
	w, h := s.Size()
	var res strings.Builder

	for y := 0; y < h; y++ {
		x := 0
		for x < w {
			r, _, _, width := s.GetContent(x, y)
			if r == 0 {
				// 空单元格，输出空格
				res.WriteRune(' ')
				x++
			} else {
				// 输出字符
				res.WriteRune(r)
				// 宽字符（如中文）占用多列，跳过占位符单元格
				// 不写入占位符，让快照文件更简洁
				x += width
			}
		}
		if y < h-1 {
			res.WriteRune('\n')
		}
	}
	return res.String()
}

// compareByDisplayWidth 使用显示宽度比较两个字符串
// 这样可以正确处理宽字符（如中文）的情况
// 文本文件无法精确表示终端单元格布局，所以使用显示宽度比较更准确
func compareByDisplayWidth(expected, actual string) bool {
	expLines := strings.Split(expected, "\n")
	actLines := strings.Split(actual, "\n")

	if len(expLines) != len(actLines) {
		return false
	}

	for i := 0; i < len(expLines); i++ {
		expLine := expLines[i]
		actLine := actLines[i]

		// 计算每行的显示宽度
		expWidth := runewidth.StringWidth(expLine)
		actWidth := runewidth.StringWidth(actLine)

		if expWidth != actWidth {
			return false
		}

		// 如果显示宽度相同，逐字符比较（考虑显示宽度）
		// 这样可以确保字符内容和对齐都正确
		expRunes := []rune(expLine)
		actRunes := []rune(actLine)

		expPos := 0
		actPos := 0
		expCol := 0
		actCol := 0

		for expPos < len(expRunes) && actPos < len(actRunes) {
			expR := expRunes[expPos]
			actR := actRunes[actPos]

			expW := runewidth.RuneWidth(expR)
			actW := runewidth.RuneWidth(actR)

			// 如果字符不同，比较失败
			if expR != actR {
				return false
			}

			expPos++
			actPos++
			expCol += expW
			actCol += actW
		}

		// 检查是否都处理完了，并且列位置对齐
		if expPos < len(expRunes) || actPos < len(actRunes) || expCol != actCol {
			return false
		}
	}

	return true
}

func assertSnapshot(t *testing.T, screen tcell.SimulationScreen, snapshotName string) {
	t.Helper()

	content := getScreenContent(screen)
	snapshotPath := filepath.Join("testdata", "snapshots", snapshotName+".txt")

	// 如果环境变量 REGO_UPDATE_SNAPSHOTS 为 true，则更新快照
	if os.Getenv("REGO_UPDATE_SNAPSHOTS") == "true" {
		err := os.MkdirAll(filepath.Dir(snapshotPath), 0755)
		if err != nil {
			t.Fatalf("Failed to create snapshot directory: %v", err)
		}
		err = os.WriteFile(snapshotPath, []byte(content), 0644)
		if err != nil {
			t.Fatalf("Failed to write snapshot: %v", err)
		}
		t.Logf("Updated snapshot: %s", snapshotPath)
		return
	}

	// 读取现有快照
	expected, err := os.ReadFile(snapshotPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatalf("Snapshot %s does not exist. Run with REGO_UPDATE_SNAPSHOTS=true to create it.", snapshotPath)
		}
		t.Fatalf("Failed to read snapshot: %v", err)
	}

	expectedStr := string(expected)

	// 使用显示宽度比较，而不是字符位置比较
	// 这样可以正确处理宽字符（如中文）的情况
	if !compareByDisplayWidth(expectedStr, content) {
		t.Errorf("Snapshot mismatch for %s\n\nExpected:\n%s\n\nGot:\n%s",
			snapshotName, expectedStr, content)
	}
}

// =============================================================================
// Snapshot 测试
// =============================================================================

func TestTextInput_Snapshot_Empty(t *testing.T) {
	app := func(c C) Node {
		return TextInput(c.Child("input"), TextInputProps{
			Placeholder: "请输入...",
			Width:       30,
		})
	}

	screen := newTestScreen(40, 10)
	tr := NewTestRuntime(app, screen)
	tr.Render()
	assertSnapshot(t, screen, "text_input_empty")
}

func TestTextInput_Snapshot_WithValue(t *testing.T) {
	app := func(c C) Node {
		return TextInput(c.Child("input"), TextInputProps{
			Value: "Hello World",
			Width: 30,
		})
	}

	screen := newTestScreen(40, 10)
	tr := NewTestRuntime(app, screen)
	tr.Render()
	assertSnapshot(t, screen, "text_input_with_value")
}

func TestTextInput_Snapshot_WithLabel(t *testing.T) {
	app := func(c C) Node {
		return TextInput(c.Child("input"), TextInputProps{
			Label:       "用户名",
			Value:       "admin",
			Placeholder: "请输入用户名",
			Width:       30,
		})
	}

	screen := newTestScreen(40, 10)
	tr := NewTestRuntime(app, screen)
	tr.Render()
	assertSnapshot(t, screen, "text_input_with_label")
}

func TestTextInput_Snapshot_Password(t *testing.T) {
	app := func(c C) Node {
		return TextInput(c.Child("input"), TextInputProps{
			Value:    "secret123",
			Password: true,
			Width:    30,
		})
	}

	screen := newTestScreen(40, 10)
	tr := NewTestRuntime(app, screen)
	tr.Render()
	assertSnapshot(t, screen, "text_input_password")
}

func TestTextInput_Snapshot_Chinese(t *testing.T) {
	app := func(c C) Node {
		return TextInput(c.Child("input"), TextInputProps{
			Value: "你好世界",
			Width: 30,
		})
	}

	screen := newTestScreen(40, 10)
	tr := NewTestRuntime(app, screen)
	tr.Render()
	assertSnapshot(t, screen, "text_input_chinese")
}

func TestTextInput_Snapshot_Multiline(t *testing.T) {
	app := func(c C) Node {
		return TextInput(c.Child("input"), TextInputProps{
			Value:     "Line 1\nLine 2\nLine 3",
			Multiline: true,
			Width:     30,
			Height:    6,
		})
	}

	screen := newTestScreen(40, 10)
	tr := NewTestRuntime(app, screen)
	tr.Render()
	assertSnapshot(t, screen, "text_input_multiline")
}
