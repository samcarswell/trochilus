package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func SetDefaultSlogLogger(l *slog.Logger) {
	slog.SetDefault(l)
}

func GetSlogHandlerOptions() *slog.HandlerOptions {
	return &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
}

func SetDefaultSlogLoggerInit(logJson bool) {
	opts := GetSlogHandlerOptions()
	if !logJson {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, opts)))
	} else {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, opts)))
	}
}

func LogErrorAndExit(logger *slog.Logger, errs ...error) {
	logger.Error(errors.Join(errs...).Error())
	os.Exit(1)
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CreateSyslog(logDir string) (*os.File, error) {
	f, err := os.Create(path.Join(logDir, "trocsys_"+randStringRunes(5)+"_"+time.Now().UTC().Format("20060102T150405")+".log"))
	if err != nil {
		return nil, err
	}

	logFile, err := os.OpenFile(f.Name(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}
	return logFile, nil
}

func PrintJson(data any) {
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		LogErrorAndExit(slog.Default(), err)
	}
	fmt.Println(string(jsonData))
}

type OutputTable[T any] struct {
	Table   table.Writer
	Rows    []T
	ConvRow func(T, OutputFormat) table.Row
}

func NewTable[T any](
	rows []T,
	convRow func(T, OutputFormat) table.Row,
	headers []string,
) OutputTable[T] {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.Style{
		Box: table.BoxStyle{
			PaddingRight: "  ",
		},
		Format: table.FormatOptions{
			Header: text.FormatDefault,
		},
	})
	var row = table.Row{}
	for _, header := range headers {
		row = append(row, header)
	}
	t.AppendHeader(row)
	return OutputTable[T]{
		Table:   t,
		Rows:    rows,
		ConvRow: convRow,
	}
}

func (t OutputTable[T]) Print(format OutputFormat) {
	if format == FormatJson {
		PrintJson(t.Rows)
	} else {
		for _, row := range t.Rows {
			t.Table.AppendRow(t.ConvRow(row, format))
		}
		switch format {
		case FormatPretty:
			t.Table.Render()
		case FormatHtml:
			t.Table.RenderHTML()
		case FormatCsv:
			t.Table.RenderCSV()
		case FormatTsv:
			t.Table.RenderTSV()
		}
	}
}

type OutputFormat string

const (
	FormatPretty OutputFormat = "pretty"
	FormatJson   OutputFormat = "json"
	FormatCsv    OutputFormat = "csv"
	FormatTsv    OutputFormat = "tsv"
	FormatHtml   OutputFormat = "html"
)

const RunAttr = "run_id"
const JobAttr = "job_name"
const EventAttr = "event"
const RunStatusAttr = "run_status"
const RunPidAttr = "run_pid"

type Event string

const EventRunCreated Event = "run-created"
const EventRunCompleted Event = "run-completed"
const EventRunStarted Event = "run-started"
const EventRunTerminated Event = "run-terminated"
const EventRunSigterm Event = "run-sigterm"

func LogRunId(runId int64) slog.Attr {
	return slog.Int64(RunAttr, runId)
}

func LogJobName(jobName string) slog.Attr {
	return slog.String(JobAttr, jobName)
}

func LogEvent(event Event) slog.Attr {
	return slog.String(EventAttr, string(event))
}

func LogRunStatus(status RunStatus) slog.Attr {
	return slog.String(RunStatusAttr, string(status))
}

func LogRunPid(pid int) slog.Attr {
	return slog.Int(RunPidAttr, pid)
}

func LogRunCreated(
	logger *slog.Logger,
	runId int64,
	jobName string,
) {
	logger.Info(
		"Run created",
		LogEvent(EventRunCreated),
		LogRunId(runId),
		LogJobName(jobName),
	)
}

func LogRunCompleted(
	logger *slog.Logger,
	runId int64,
	jobName string,
	status RunStatus,
) {
	logger.Info(
		"Run completed",
		LogEvent(EventRunCompleted),
		LogRunId(runId),
		LogJobName(jobName),
		LogRunStatus(status),
	)
}

func LogRunStarted(
	logger *slog.Logger,
	runId int64,
	jobName string,
	pid int,
) {
	logger.Info(
		"Run started",
		LogEvent(EventRunStarted),
		LogRunId(runId),
		LogJobName(jobName),
		LogRunPid(pid),
	)
}

func LogRunSentSigterm(
	logger *slog.Logger,
	runId int64,
	jobName string,
	pid int,
) {
	logger.Info(
		"Run sent SIGTERM",
		LogEvent(EventRunSigterm),
		LogRunId(runId),
		LogJobName(jobName),
		LogRunPid(pid),
	)
}

func LogRunTerminated(
	logger *slog.Logger,
	runId int64,
	jobName string,
	signalErr string,
) {
	logger.Error(
		"Run has been terminated: "+signalErr,
		LogEvent(EventRunTerminated),
		LogRunId(runId),
		LogJobName(jobName),
	)
}
