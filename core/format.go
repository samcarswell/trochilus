package core

// TODO: need to fix this up for status
func FormatSucceeded(succeeded bool) string {
	if succeeded {
		return "✅"
	} else {
		return "❌"
	}
}
