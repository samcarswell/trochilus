package core

type RunStatus string

const (
	RunStatusRunning   RunStatus = "Running"
	RunStatusSkipped   RunStatus = "Skipped"
	RunStatusSucceeded RunStatus = "Succeeded"
	RunStatusFailed    RunStatus = "Failed"
)
