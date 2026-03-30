package notify

import (
	"testing"

	"github.com/samcarswell/trochilus/config"
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
		tagStatuses      config.StatusConfig
	}{
		{"success", "test-1", 34, core.RunStatusSucceeded, "/file/path", false, "server1.com",
			`*test-1@server1.com*: run 34 - ✅ Succeeded
Log: ` + "`/file/path`", true, config.StatusConfig{}},
		{"failed", "test-2", 10000000, core.RunStatusFailed, "/file/path", false, "server1.com",
			`*test-2@server1.com*: run 10000000 - ❌ Failed <!channel>
Log: ` + "`/file/path`", true, config.StatusConfig{Failed: true}},
		{"no-log", "test-3", 34, core.RunStatusSucceeded, "", false, "server1.com",
			`*test-3@server1.com*: run 34 - ✅ Succeeded`, true, config.StatusConfig{}},
		{"empty-hostname", "test-4", 34, core.RunStatusSucceeded, "/file/path", false, "",
			`*test-4*: run 34 - ✅ Succeeded
Log: ` + "`/file/path`", true, config.StatusConfig{}},
		{"notify-log-content", "test-5", 34, core.RunStatusSucceeded, "testdata/example.log", true, "",
			`*test-5*: run 34 - ✅ Succeeded
Log:
` + "```" + "\nLine one of log\nLine two of log\n```", true, config.StatusConfig{}},
		{"notify-log-content-file-does-not-exist", "test-6", 34, core.RunStatusSucceeded, "testdata/notreal.log", true, "", "*test-6*: run 34 - ✅ Succeeded", true, config.StatusConfig{}},
		{"success", "test-7", 34, core.RunStatusSucceeded, "/file/path", false, "server1.com",
			`*test-7@server1.com*: run 34 - Succeeded
Log: ` + "`/file/path`", false, config.StatusConfig{}},
		{"success-tag-config", "test-8", 10000000, core.RunStatusSucceeded, "/file/path", false, "server1.com",
			`*test-8@server1.com*: run 10000000 - ✅ Succeeded <!channel>
Log: ` + "`/file/path`", true, config.StatusConfig{Succeeded: true}},
		{"skipped-tag-config", "test-9", 10000000, core.RunStatusSkipped, "/file/path", false, "server1.com",
			`*test-9@server1.com*: run 10000000 - ⚠️ Skipped <!channel>
Log: ` + "`/file/path`", true, config.StatusConfig{Skipped: true}},
		{"running-tag-config", "test-10", 10000000, core.RunStatusRunning, "/file/path", false, "server1.com",
			`*test-10@server1.com*: run 10000000 - 🚀 Running <!channel>
Log: ` + "`/file/path`", true, config.StatusConfig{Running: true}},
		{"terminated-tag-config", "test-11", 10000000, core.RunStatusTerminated, "/file/path", false, "server1.com",
			`*test-11@server1.com*: run 10000000 - 💥 Terminated <!channel>
Log: ` + "`/file/path`", true, config.StatusConfig{Terminated: true}},
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
				d.tagStatuses,
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
