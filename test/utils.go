package test

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math/rand"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
	"github.com/samcarswell/trochilus/data"
)

var (
	_, b, _, _  = runtime.Caller(0)
	basepath    = filepath.Dir(b)
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
)

type logRow struct {
	Time  time.Time `json:"time"`
	Level string    `json:"level"`
	Msg   string    `json:"msg"`
}

func MigrationsDir() string {
	return path.Join(basepath, "..", "db", "migrations")
}

func CreateDb(ctx context.Context, t *testing.T) *data.Queries {
	dir := t.TempDir()
	dbPath := path.Join(dir, "troc.db")
	t.Log("Creating database at " + dbPath)
	f := os.DirFS(MigrationsDir())
	return data.New(config.CreateOrUpdateDatabase(
		f,
		context.Background(),
		dbPath,
		".", // Not sure why this works. Passing the correct path doesn't work
	))
}

func CreateSysLogFile(t *testing.T) (string, *slog.Logger) {
	logFile, err := core.CreateSyslog(t.TempDir())
	if err != nil {
		t.Error(err)
	}
	t.Log("Created syslog file at " + logFile.Name())
	return logFile.Name(), CreateJsonLogger(logFile)
}

func UniqueIdentifer() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CreateJsonLogger(logFile *os.File) *slog.Logger {
	return slog.New(slog.NewJSONHandler(io.MultiWriter(logFile, os.Stderr), &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}

func AssertInt[V int | int8 | int16 | int32 | int64](t *testing.T, expected V, actual V) {
	if actual != expected {
		t.Fatalf("incorrect result: expected %d, got %d", expected, actual)
	}
}

func AssertString(t *testing.T, expected string, actual string) {
	if actual != expected {
		t.Fatalf("incorrect result: expected %s, got %s", expected, actual)
	}
}

func AssertBool(t *testing.T, expected bool, actual bool) {
	if actual != expected {
		t.Fatalf("incorrect result: expected %t, got %t", expected, actual)
	}
}

func AssertFileExists(t *testing.T, path string) {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file does not exist at %s", path)
	}
}

func AssertFileContents(t *testing.T, expected string, path string) {
	file, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err.Error())
	}
	AssertString(t, expected, string(file))
}

func AssertLogHasInfo(t *testing.T, text string, path string) {
	assertLogHasLine(t, "INFO", text, path)
}

func AssertLogHasWarn(t *testing.T, text string, path string) {
	assertLogHasLine(t, "WARN", text, path)
}

func AssertLogDoesNotHaveInfo(t *testing.T, text string, path string) {
	assertLogDoesNotHaveLine(t, "INFO", text, path)
}

func assertLogHasLine(t *testing.T, level string, text string, path string) {
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var row logRow
		err := json.Unmarshal([]byte(scanner.Text()), &row)
		if err != nil {
			t.Fatal(err.Error())
		}
		if row.Level == level && row.Msg == text {
			return
		}
	}

	t.Fatalf("log does not contain line with level %s and message: %s", level, text)
}

func assertLogDoesNotHaveLine(t *testing.T, level string, text string, path string) {
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err.Error())
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var row logRow
		err := json.Unmarshal([]byte(scanner.Text()), &row)
		if err != nil {
			t.Fatal(err.Error())
		}
		if row.Level == level && row.Msg == text {
			t.Fatalf("log contains line with level %s and message: %s", level, text)
		}
	}
}
