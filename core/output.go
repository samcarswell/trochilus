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

func GetDefaultTextSlogLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, GetSlogHandlerOptions()))
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
