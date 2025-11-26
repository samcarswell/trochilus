package core

import (
	"database/sql"
	"strconv"
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

func FormatPid(value sql.NullInt64) string {
	if value.Valid {
		return strconv.FormatInt(value.Int64, 10)
	}
	return ""
}

func FormatDuration(start time.Time, end time.Time) string {
	if end.IsZero() {
		return ""
	}
	return end.Sub(start).String()
}
