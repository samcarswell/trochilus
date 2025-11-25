package core

import (
	"time"
)

func FormatStatus(status RunStatus, showEmoji bool) string {
	switch status {
	case RunStatusSucceeded:
		return formatEmoji("✅", showEmoji) + string(status)
	case RunStatusFailed:
		return formatEmoji("❌", showEmoji) + string(status)
	case RunStatusRunning:
		return formatEmoji("🚀", showEmoji) + string(status)
	case RunStatusSkipped:
		return formatEmoji("⚠️", showEmoji) + string(status)
	case RunStatusTerminated:
		return formatEmoji("💥", showEmoji) + string(status)
	}
	return string(status)
}

func formatEmoji(emoji string, showEmoji bool) string {
	if showEmoji {
		return emoji + " "
	}
	return ""
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
