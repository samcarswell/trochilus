package config

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"log"
	"log/slog"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/samcarswell/trochilus/core"
	"github.com/samcarswell/trochilus/data"
	"github.com/spf13/viper"
	_ "modernc.org/sqlite"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
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

func CreateAndReadConfig(
	confDir string,
	confName string,
	confType string,
) {
	expandedConfigDir, err := expandDir(confDir)
	if err != nil {
		log.Fatalf("Unable to expand configuration directory %s %s", confDir, err)
	}
	err = viper.ReadInConfig()
	if err != nil {
		var confNotFoundErr viper.ConfigFileNotFoundError
		if errors.As(err, &confNotFoundErr) {
			log.Println("Creating config directory at " + expandedConfigDir)
			err := os.MkdirAll(expandedConfigDir, os.ModePerm)
			if err != nil {
				log.Fatalf("Unable to create config directory: %s %s", expandedConfigDir, err)
			}
			log.Println("Creating initial config file at " + expandedConfigDir + "/" + confName + "." + confType)
			err = viper.SafeWriteConfig()
			if err != nil {
				log.Fatalf("Unable to write initial config: %s", err)
			}
		} else {
			log.Fatalf("Unable to read config: %s", err)
		}
	}
}

func CreateOrUpdateDatabase(
	migrations fs.FS,
	ctx context.Context,
	dbPath string,
	migrationsDir string,
) *sql.DB {
	fileName := path.Base(dbPath)
	dir := path.Dir(dbPath)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Fatalf("Unable to create database path %s", err)
	}
	u, _ := url.Parse("sqlite3:///" + dbPath)
	dbMateDb := dbmate.New(u)
	dbMateDb.FS = migrations
	dbMateDb.AutoDumpSchema = false
	dbMateDb.MigrationsDir = []string{migrationsDir}
	dbMateDb.Log = os.Stderr

	err = dbMateDb.CreateAndMigrate()
	if err != nil {
		log.Fatalf("Unable to update database %s", err)
	}
	db, err := sql.Open("sqlite", path.Join(dir, fileName)+"?mode=rw")
	if err != nil {
		log.Fatalf("Unable to open database %s", err)
	}
	db.Exec("PRAGMA journal_mode=WAL;")
	return db
}

func GetDatabase(ctx context.Context) *data.Queries {
	migrations, ok := MigrationsFromContext(ctx)
	if !ok {
		log.Fatalf("Could not get migrations")
	}
	dbPath := viper.GetString(ConfigDatabasePath)
	if dbPath == "" {
		log.Fatalln("database config value is empty")
	}
	expandedPath, err := expandDir(dbPath)
	if err != nil {
		log.Fatalf("Unable to expand database path %s", err)
	}

	return data.New(CreateOrUpdateDatabase(
		migrations,
		ctx,
		expandedPath,
		"./db/migrations",
	))
}

func GetLoggerOrExit(ctx context.Context) *slog.Logger {
	logger, ok := LoggerFromContext(ctx)
	if !ok {
		log.Fatalln("Could not get logger from context")
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

type NotifyConfig struct {
	Hostname string
	Slack    SlackConfig
}

type SlackConfig struct {
	Token   string
	Channel string
}

type Config struct {
	Database string
	LockDir  string
	LogDir   string
	Notify   NotifyConfig
}

func GetConfig() Config {
	return Config{
		Database: viper.GetString("database"),
		LockDir:  viper.GetString("lockdir"),
		LogDir:   viper.GetString("logdir"),
		Notify: NotifyConfig{
			Hostname: viper.GetString("notify.hostname"),
			Slack: SlackConfig{
				Token:   viper.GetString("notify.slack.token"),
				Channel: viper.GetString("notify.slack.channel"),
			},
		},
	}
}

type loggerKey struct{}
type logFileKey struct{}
type migrationsKey struct{}

func ContextWithLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

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

func ContextWithMigrations(ctx context.Context, migrations embed.FS) context.Context {
	return context.WithValue(ctx, migrationsKey{}, migrations)
}

func MigrationsFromContext(ctx context.Context) (embed.FS, bool) {
	migrations, ok := ctx.Value(migrationsKey{}).(embed.FS)
	return migrations, ok
}
