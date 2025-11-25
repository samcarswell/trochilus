package notify

import (
	"testing"

	"github.com/samcarswell/trochilus/core"
)

func Test_getNotifyText(t *testing.T) {
	data := []struct {
		name             string
		jobName          string
		runId            int64
		runStatus        core.RunStatus
		logFile          string
		notifyLogContent bool
		hostname         string
		expected         string
		showEmoji        bool
	}{
		{"success", "test-1", 34, core.RunStatusSucceeded, "/file/path", false, "server1.com",
			`*test-1@server1.com*: run 34 - ✅ Succeeded
Log: ` + "`/file/path`", true},
		{"failed", "test-2", 10000000, core.RunStatusFailed, "/file/path", false, "server1.com",
			`*test-2@server1.com*: run 10000000 - ❌ Failed <!channel>
Log: ` + "`/file/path`", true},
		{"no-log", "test-3", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-3@server1.com*: run 34 - ✅ Succeeded`, true},
		{"empty-hostname", "test-4", 34, core.RunStatusSucceeded, "/file/path", false, "",
			`*test-4*: run 34 - ✅ Succeeded
Log: ` + "`/file/path`", true},
		{"notify-log-content", "test-5", 34, core.RunStatusSucceeded, "testdata/example.log", true, "",
			`*test-5*: run 34 - ✅ Succeeded
Log:
` + "```" + "\nLine one of log\nLine two of log\n```", true},
		{"notify-log-content-file-does-not-exist", "test-6", 34, core.RunStatusSucceeded, "testdata/notreal.log", true, "", "*test-6*: run 34 - ✅ Succeeded", true},
		{"success", "test-7", 34, core.RunStatusSucceeded, "/file/path", false, "server1.com",
			`*test-7@server1.com*: run 34 - Succeeded
Log: ` + "`/file/path`", false},
	}

	for _, d := range data {
		t.Run(d.name, func(t *testing.T) {
			notifyStr := getNotifyText(
				RunNotifyInfo{
					Name:             d.jobName,
					NotifyLogContent: d.notifyLogContent,
					Id:               d.runId,
					Status:           d.runStatus,
					LogFile:          d.logFile,
				},
				d.hostname,
				d.showEmoji,
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
