/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"io"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"

	"carswellpress.com/trochilus/config"
	"carswellpress.com/trochilus/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cliName = "troc"
var configPath = "$HOME/.config/troc"
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
		log.Fatalln("Unable to get logdir from config %w", err)
	}
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Fatalln("Unable to create logdir %w", err)
	}

	var l *slog.Logger
	if cmd.CommandPath() == cliName+" exec" {
		// If we're executing a cron, we need to log to file
		logFile, err := core.CreateSyslog(logDir)
		if err != nil {
			log.Fatalln("Unable to create trocsys log %w", err)
		}
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

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("Unable to read hostname. Defaulting to unknown-hostname")
		hostname = "unknown-hostname"
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Unable to get user home dir %s", err)
	}
	viper.SetDefault("database", path.Join(homedir, ".config", "troc", "troc.db"))
	viper.SetDefault("logdir", os.TempDir())
	viper.SetDefault("lockdir", os.TempDir())
	viper.SetDefault("hostname", hostname)
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AutomaticEnv()
	config.CreateAndReadConfig(configPath, configName, configType)
}
