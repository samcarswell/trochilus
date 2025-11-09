/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"embed"
	"log"
	"log/slog"
	"os"
	"path"
	"strings"

	slogmulti "github.com/samber/slog-multi"
	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cliName = "troc"
var configPath = "$HOME/.config/troc"
var configName = "config"
var configType = "yaml"
var version = "development"

var RootCmd = &cobra.Command{
	Use:     cliName,
	Version: version,
	Short:   "Trochilus - simple cron monitoring",
	Long: `Trochilus - simple cron monitoring
	
https://github.com/samcarswell/trochilus
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupContext(cmd)
	},
}

func setupContext(cmd *cobra.Command) {
	conf := config.GetConfig()
	err := os.MkdirAll(conf.LogDir, os.ModePerm)
	if err != nil {
		log.Fatalln("Unable to create logdir %w", err)
	}
	var l *slog.Logger
	if cmd.CommandPath() == cliName+" exec" {
		// If we're executing a cron, we need to log to file
		logFile, err := core.CreateSyslog(conf.LogDir)
		if err != nil {
			log.Fatalln("Unable to create trocsys log %w", err)
		}
		l = slog.New(slogmulti.Fanout(
			slog.Default().Handler(),
			slog.NewJSONHandler(logFile, core.GetSlogHandlerOptions()),
		))
		l.Info("Logging to " + logFile.Name())
		cmd.SetContext(config.ContextWithLogFile(cmd.Context(), logFile.Name()))
	} else {
		l = slog.Default()
	}

	cmd.SetContext(config.ContextWithLogger(cmd.Context(), l))
	cmd.SetContext(config.ContextWithMigrations(cmd.Context(), Migrations))
}

var Migrations embed.FS

func Execute(migrations embed.FS) {
	Migrations = migrations
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	core.SetDefaultSlogLogger()
	RootCmd.SetHelpCommand(&cobra.Command{Hidden: true})
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
	viper.SetDefault("notify.hostname", hostname)
	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AutomaticEnv()
	config.CreateAndReadConfig(configPath, configName, configType)
}
