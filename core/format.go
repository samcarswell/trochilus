package core

func FormatSucceeded(succeeded bool) string {
	if succeeded {
		return "✅"
	} else {
		return "❌"
	}
}
