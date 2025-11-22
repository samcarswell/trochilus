package core

type RunStatus string

const (
	RunStatusRunning    RunStatus = "Running"
	RunStatusSkipped    RunStatus = "Skipped"
	RunStatusSucceeded  RunStatus = "Succeeded"
	RunStatusFailed     RunStatus = "Failed"
	RunStatusTerminated RunStatus = "Terminated"
)

type RunShow struct {
	ID            int64
	CronName      string
	StartTime     string
	EndTime       string
	LogFile       string
	SystemLogFile string
	Status        string
	Duration      string
}
