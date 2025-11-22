package core

import (
	"time"
)

func FormatStatus(status RunStatus) string {
	switch status {
	case RunStatusSucceeded:
		return "✅ " + string(status)
	case RunStatusFailed:
		return "❌ " + string(status)
	case RunStatusRunning:
		return "⚙️ " + string(status)
	case RunStatusSkipped:
		return "⚠️ " + string(status)
	case RunStatusKilled:
		return "💀 " + string(status)
	}
	return string(status)
}

func FormatTime(timeValue time.Time, useLocalTime bool) string {
	if timeValue.IsZero() {
		return ""
	}
	if useLocalTime {
		return timeValue.In(time.Local).String()
	}
	return timeValue.String()
}
