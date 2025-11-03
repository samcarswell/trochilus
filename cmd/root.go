/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"io"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"carswellpress.com/trochilus/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cliName = "troc"
var configPath = "$HOME/.config/trochilus"
var configName = "config"
var configType = "yaml"

var RootCmd = &cobra.Command{
	Use:   cliName,
	Short: "Trochilus - simple cron monitoring",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupContext(cmd)
	},
}

func setupContext(cmd *cobra.Command) {
	logDir, err := config.GetLogDir()
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		panic(err)
	}

	var l *slog.Logger
	if cmd.CommandPath() == cliName+" exec" {
		// If we're executing a cron, we need to log to file
		f, err := os.Create(path.Join(logDir, "cc_"+time.Now().UTC().Format("20060102T150405")+".log"))
		if err != nil {
			panic(err)
		}

		logFile, _ := os.OpenFile(f.Name(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		l = slog.New(slog.NewTextHandler(io.MultiWriter(logFile, os.Stdout), &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
		l.Info("Logging to " + logFile.Name())
		cmd.SetContext(config.ContextWithLogFile(cmd.Context(), logFile.Name()))
	} else {
		l = slog.Default()
	}

	cmd.SetContext(config.ContextWithLogger(cmd.Context(), l))
	cmd.SetContext(config.ContextWithSchema(cmd.Context(), SqlSchema))
}

var SqlSchema string

func Execute(sqlSchema string) {
	SqlSchema = sqlSchema
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.SetEnvPrefix("TROC")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetDefault("database", "$HOME/.config/trochilus/cc.db")
	viper.SetDefault("logdir", "$TMPDIR")
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AutomaticEnv()
	config.CreateAndReadConfig(configPath, configName, configType)
}
