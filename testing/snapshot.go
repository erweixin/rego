package testing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// AssertSnapshot 比较当前屏幕内容与快照文件
func AssertSnapshot(t *testing.T, screen *MockScreen, snapshotName string) {
	t.Helper()

	content := screen.GetContentString()
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
	if content != expectedStr {
		t.Errorf("Snapshot mismatch for %s\n\nExpected:\n%s\n\nGot:\n%s\n\nDiff:\n%s",
			snapshotName, expectedStr, content, diff(expectedStr, content))
	}
}

// 简单的 diff 实现
func diff(expected, actual string) string {
	expLines := strings.Split(expected, "\n")
	actLines := strings.Split(actual, "\n")

	var res strings.Builder
	maxLines := len(expLines)
	if len(actLines) > maxLines {
		maxLines = len(actLines)
	}

	for i := 0; i < maxLines; i++ {
		var expLine, actLine string
		if i < len(expLines) {
			expLine = expLines[i]
		}
		if i < len(actLines) {
			actLine = actLines[i]
		}

		if expLine != actLine {
			res.WriteString("-" + expLine + "\n")
			res.WriteString("+" + actLine + "\n")
		} else {
			res.WriteString(" " + expLine + "\n")
		}
	}
	return res.String()
}

