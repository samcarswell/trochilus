package core

// TODO: need to fix this up for status
func FormatStatus(status RunStatus) string {
	switch status {
	case RunStatusSucceeded:
		return string(status) + ": ✅"
	case RunStatusFailed:
		return string(status) + ": ❌"
	case RunStatusRunning:
		return string(status) + ": ⚙️"
	case RunStatusSkipped:
		return string(status) + ": ⚠️"
	}
	return string(status)
}
