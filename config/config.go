package config

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"path"
	"strings"

	"carswellpress.com/trochilus/core"
	"carswellpress.com/trochilus/data"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"
)

const ConfigDatabasePath = "database"
const ConfigLogDir = "logdir"

func expandDir(path string) (string, error) {
	if strings.HasSuffix(path, "~") {
		path = strings.Replace(path, "~", "$HOME", 1)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	tmpDir := os.TempDir()

	mapper := func(placeholderName string) string {
		switch placeholderName {
		case "HOME":
			return homeDir
		case "TMPDIR":
			return tmpDir
		}
		return ""
	}
	return os.Expand(path, mapper), nil
}

func GetLogDir() (string, error) {
	logDir := viper.GetString(ConfigLogDir)
	expandedDir, err := expandDir(logDir)
	if err != nil {
		return "", err
	}
	return expandedDir, nil
}

func GetDatabase(ctx context.Context) *data.Queries {
	dbPath := viper.GetString(ConfigDatabasePath)
	if dbPath == "" {
		panic("database config value is empty")
	}
	expandedPath, err := expandDir(dbPath)
	if err != nil {
		panic(err)
	}
	fileName := path.Base(expandedPath)
	dir := path.Dir(expandedPath)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		panic(err)
	}
	db, err := sql.Open("sqlite", path.Join(dir, fileName)+"?mode=rw")
	if err != nil {
		panic(err)
	}

	schema, ok := SchemaFromContext(ctx)
	if !ok {
		panic("Could not get schema")
	}

	if _, err := db.ExecContext(context.Background(), schema); err != nil {
		panic(err.Error())
	}

	return data.New(db)
}

func GetLoggerOrExit(ctx context.Context) *slog.Logger {
	logger, ok := LoggerFromContext(ctx)
	if !ok {
		// Should this be a panic?
		panic("Could not get logger from context")
	}
	return logger
}

func GetLogFileOrExit(logger *slog.Logger, ctx context.Context) string {
	logFile, ok := LogFileFromContext(ctx)
	if !ok {
		core.LogErrorAndExit(logger, errors.New("Could not get logFile from context"))
	}
	return logFile
}

func GetSlackToken() string {
	return viper.GetString("slack.token")
}

func GetSlackChannel() string {
	return viper.GetString("slack.channel")
}

type loggerKey struct{}
type logFileKey struct{}
type schemaKey struct{}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// TODO: the logger dones't always seem to send all logs.
func LoggerFromContext(ctx context.Context) (*slog.Logger, bool) {
	dbConn, ok := ctx.Value(loggerKey{}).(*slog.Logger)
	return dbConn, ok
}

func ContextWithLogFile(ctx context.Context, logFile string) context.Context {
	return context.WithValue(ctx, logFileKey{}, logFile)
}

func LogFileFromContext(ctx context.Context) (string, bool) {
	logFile, ok := ctx.Value(logFileKey{}).(string)
	return logFile, ok
}

func ContextWithSchema(ctx context.Context, schema string) context.Context {
	return context.WithValue(ctx, schemaKey{}, schema)
}

func SchemaFromContext(ctx context.Context) (string, bool) {
	schema, ok := ctx.Value(schemaKey{}).(string)
	return schema, ok
}
