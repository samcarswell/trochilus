package notify

import (
	"testing"

	"github.com/samcarswell/trochilus/core"
)

func Test_getNotifyText(t *testing.T) {
	data := []struct {
		name             string
		cronName         string
		runId            int64
		runStatus        core.RunStatus
		logFile          string
		notifyLogContent bool
		hostname         string
		expected         string
	}{
		{"success", "test-1", 34, core.RunStatusSucceeded, "/file/path", false, "server1.com",
			`*test-1@server1.com*: run 34 - ✅ Succeeded
Log: ` + "`/file/path`"},
		{"failed", "test-2", 10000000, core.RunStatusFailed, "/file/path", false, "server1.com",
			`*test-2@server1.com*: run 10000000 - ❌ Failed <!channel>
Log: ` + "`/file/path`"},
		{"no-log", "test-3", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-3@server1.com*: run 34 - ✅ Succeeded`},
		{"empty-hostname", "test-3", 34, core.RunStatusSucceeded, "/file/path", false, "",
			`*test-3*: run 34 - ✅ Succeeded
Log: ` + "`/file/path`"},
		{"notify-log-content", "test-3", 34, core.RunStatusSucceeded, "testdata/example.log", true, "",
			`*test-3*: run 34 - ✅ Succeeded
Log: ` + "`testdata/example.log`" + "\nLog Content:\n" + "```" + "\nLine one of log\nLine two of log\n```"},
		{"notify-log-content-file-does-not-exist", "test-3", 34, core.RunStatusSucceeded, "testdata/notreal.log", true, "",
			`*test-3*: run 34 - ✅ Succeeded
Log: ` + "`testdata/notreal.log`"},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			notifyStr := getNotifyText(
				RunNotifyInfo{
					Name:             d.cronName,
					NotifyLogContent: d.notifyLogContent,
					Id:               d.runId,
					Status:           d.runStatus,
					LogFile:          d.logFile,
				},
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
