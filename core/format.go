package core

import "time"

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
	}
	return string(status)
}

func FormatTime(timeValue time.Time) string {
	if timeValue.IsZero() {
		return ""
	}
	return timeValue.String()
}
