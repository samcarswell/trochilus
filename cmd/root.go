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
	"time"

	"carswellpress.com/trochilus/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cliName = "troc"

// RootCmd represents the base command when called without any subcommands
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

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(sqlSchema string) {
	SqlSchema = sqlSchema
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cron-cowboy.yaml)")
	viper.SetEnvPrefix("TROC")
	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)

	viper.SetDefault("database", "$HOME/.config/trochilus/cc.db")
	viper.SetDefault("logdir", "$TMPDIR")
	viper.AddConfigPath("$HOME/.config/trochilus")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Fatalf("Unable to read config: %s", err)
	}

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
