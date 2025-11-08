package notify

import (
	"testing"

	"github.com/samcarswell/trochilus/core"
)

func Test_getNotifyText(t *testing.T) {
	data := []struct {
		name      string
		cronName  string
		runId     int64
		runStatus core.RunStatus
		logFile   string
		hostname  string
		expected  string
	}{
		{"success", "test-1", 34, core.RunStatusSucceeded, "/file/path", "server1.com",
			`*test-1@server1.com*: run 34 - ✅ Succeeded
Log: ` + "`/file/path`"},
		{"failed", "test-2", 10000000, core.RunStatusFailed, "/file/path", "server1.com",
			`*test-2@server1.com*: run 10000000 - ❌ Failed <!channel>
Log: ` + "`/file/path`"},
		{"no-log", "test-3", 34, core.RunStatusSucceeded, "", "server1.com",
			`*test-3@server1.com*: run 34 - ✅ Succeeded`},
		{"empty-hostname", "test-3", 34, core.RunStatusSucceeded, "/file/path", "",
			`*test-3*: run 34 - ✅ Succeeded
Log: ` + "`/file/path`"},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			notifyStr := getNotifyText(
				d.cronName,
				d.runId,
				d.runStatus,
				d.logFile,
				d.hostname,
			)
			if notifyStr != d.expected {
				t.Error("Expected")
				t.Error(d.expected)
				t.Error("Actual")
				t.Fatal(notifyStr)
			}
		})
	}
}
