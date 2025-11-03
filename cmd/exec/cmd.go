/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"carswellpress.com/trochilus/cmd"
	"carswellpress.com/trochilus/config"
	"carswellpress.com/trochilus/core"
	"carswellpress.com/trochilus/data"
	"carswellpress.com/trochilus/notify"
	"carswellpress.com/trochilus/opts"
	"github.com/gofrs/flock"
	"github.com/spf13/cobra"
)

var nameOpt = "name"
var notifyOpt = "notify"

// While *OrExit is useful for most commands, exec actually needs to
// try it's best to recover: it should try to get least get a message to the slack channel notifying of a failure.
// For now I'm just going to assume there's no errors. But this should
// be revisited before a v1
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Run a CRON command",
	Run: func(cmd *cobra.Command, args []string) {
		logger := config.GetLoggerOrExit(cmd.Context())
		cronName := opts.GetStringOptOrExit(logger, cmd, nameOpt)
		notifyOpt := opts.GetBoolOptOrExit(logger, cmd, notifyOpt)
		slackConf := config.GetSlackConfig()
		hostname := config.GetHostnameConfig()
		queries := config.GetDatabase(cmd.Context())
		if notifyOpt && slackConf.Token == "" {
			core.LogErrorAndExit(logger, errors.New("notify is set but slack.token is blank."))
		}
		if notifyOpt && slackConf.Channel == "" {
			core.LogErrorAndExit(logger, errors.New("notify is set but slack.channel is blank."))
		}

		logFile := config.GetLogFileOrExit(logger, cmd.Context())

		if len(args) == 0 {
			core.LogErrorAndExit(logger, errors.New("Must provide args"))
		}

		cronRow, err := queries.GetCron(context.Background(), cronName)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Info("Cron not registered")
			} else {
				core.LogErrorAndExit(logger, err)
			}
		}
		if cronRow == (data.GetCronRow{}) {
			id, err := queries.CreateCron(context.Background(), cronName)
			if err != nil {
				core.LogErrorAndExit(logger, err)
			}
			cronRow.Cron.Name = cronName
			cronRow.Cron.ID = id
		}

		// TODO: lock dir should be configurable
		lockFile := filepath.Join(os.TempDir(), cronName+".lock")
		f := flock.New(lockFile)

		locked, err := f.TryLock()

		if err != nil || !locked {
			skipRun(cronRow.Cron.ID, logFile, slackConf, queries, context.Background(), logger, hostname)
		}
		if !locked {
			log.Fatalf("Unable to create lock for cron. Likely already running")
		}

		defer f.Unlock()
		logger.Info("Created cron lock at " + lockFile)

		dir, err := config.GetLogDir()
		if err != nil {
			log.Fatalf("Unable to get logdir %s", err)
		}
		stdout, err := os.CreateTemp(dir, cronName+".*.log")
		if err != nil {
			log.Fatalf("Unable to create log file %s", err)
		}
		stdoutLog, err := os.OpenFile(stdout.Name(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Fatalf("Unable to open log file %s", err)
		}

		logger.Info("Run log created at: " + stdout.Name())
		runId, err := queries.StartRun(context.Background(), data.StartRunParams{
			CronID:      cronRow.Cron.ID,
			LogFile:     stdout.Name(),
			ExecLogFile: logFile,
		})
		if err != nil {
			log.Fatalf("Unable to start run %s", err)
		}

		logger.Info("Run created with ID " + strconv.FormatInt(runId, 10))

		// time.Sleep(10 * time.Second)
		cmdArgs := strings.Split(args[0], " ")
		runCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		runCmd.Stdout = stdoutLog
		runCmd.Stderr = stdoutLog
		err = runCmd.Run()
		status := core.RunStatusSucceeded
		if err != nil {
			logger.Error("Error occurred during run", "error", err)
			status = core.RunStatusFailed
		}
		queries.EndRun(context.Background(), data.EndRunParams{
			Status: string(status),
			ID:     runId,
		})
		logger.Info("Run " + strconv.FormatInt(runId, 10) + " completed: " + string(status))

		completedRun, err := queries.GetRun(cmd.Context(), runId)
		if err != nil {
			log.Fatalf("Unable to get completed run %s", err)
		}

		if notifyOpt {
			ok, err := notify.NotifyRunSlack(slackConf, completedRun, hostname)
			if err != nil {
				log.Fatalf("Unable to notify slack %s", err)
			}
			if !ok {
				logger.Error("Command was run, but slack was unable to be notified")
			}
		}
	},
}

func init() {
	cmd.RootCmd.AddCommand(execCmd)

	execCmd.Flags().String(nameOpt, "", "Name of cron to execute. Will create it if it does not exist")
	if err := execCmd.MarkFlagRequired(nameOpt); err != nil {
		log.Fatalf("Unable to mark "+nameOpt+" as required %s", err)
	}
	execCmd.Flags().Bool(notifyOpt, false, "Notifies of the exec success")
}

func skipRun(
	cronId int64,
	execLogFile string,
	slackConf config.SlackConfig,
	queries *data.Queries,
	ctx context.Context,
	logger *slog.Logger,
	hostname string,
) {
	id, err := queries.SkipRun(ctx, data.SkipRunParams{
		CronID:      cronId,
		ExecLogFile: execLogFile,
	})
	if err != nil {
		// TODO: need a standard function here to deal with errors and communicate to slack
		log.Fatalf("Unable to skip run %s", err)
	}
	run, err := queries.GetRun(ctx, id)
	if err != nil {
		// TODO: need a standard function here to deal with errors and communicate to slack
		log.Fatalf("Unable to get updated run %s", err)
	}
	logger.Warn("Skipping run " + strconv.FormatInt(id, 10) + ". Cron is already running.")
	notify.NotifyRunSlack(slackConf, run, hostname)
	os.Exit(1)
}
