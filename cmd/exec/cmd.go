/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	"errors"
	"log"
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

// While *OrExit is useful for most commands, exec actually needs to
// try it's best to recover: it should try to get least get a message to the slack channel notifying of a failure.
// For now I'm just going to assume there's no errors. But this should
// be revisited before a v1
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Run a CRON command",
	Run: func(cmd *cobra.Command, args []string) {
		logger := config.GetLoggerOrExit(cmd.Context())
		cronName := opts.GetStringOptOrExit(logger, cmd, "cron-name")
		skipCrontab := opts.GetBoolOptOrExit(logger, cmd, "skip-crontab")
		notifyOpt := opts.GetBoolOptOrExit(logger, cmd, "notify")
		slackToken := config.GetSlackToken()
		if notifyOpt && slackToken == "" {
			core.LogErrorAndExit(logger, errors.New("notify is set but slack.token is blank."))
		}
		channel := config.GetSlackChannel()
		if notifyOpt && channel == "" {
			core.LogErrorAndExit(logger, errors.New("notify is set but slack.channel is blank."))
		}

		logFile := config.GetLogFileOrExit(logger, cmd.Context())

		if len(args) == 0 {
			core.LogErrorAndExit(logger, errors.New("Must provide args"))
		}

		crons, err := core.GetCrontabItems(logger)
		if err != nil {
			core.LogErrorAndExit(logger, err)
		}

		found := false
		for _, value := range crons {
			if value.Name == cronName {
				found = true
			}
		}
		if !found && !skipCrontab {
			core.LogErrorAndExit(logger,
				errors.New("Unable to find cron defined with name "+cronName),
			)
		}

		// TODO: lock dir should be configurable
		lockFile := filepath.Join(os.TempDir(), cronName+".lock")
		f := flock.New(lockFile)

		locked, err := f.TryLock()

		if err != nil {
			// TODO: this needs to talk to slack. We probably need another
			// type of run; skipped runs. For now I'll just exit 1
			log.Fatalf("Unable to create lock for cron. Likely already running: %s", err)
		}
		if !locked {
			log.Fatalf("Unable to create lock for cron. Likely already running")
		}

		defer f.Unlock()
		logger.Info("Created cron lock at " + lockFile)

		queries := config.GetDatabase(cmd.Context())
		cronRow, err := queries.GetCron(context.Background(), cronName)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Info("Cron not registered")
			} else {
				core.LogErrorAndExit(logger, err)
			}
		}
		if cronRow == (data.GetCronRow{}) {
			logger.Info("Registering cron")
			id, err := queries.CreateCron(context.Background(), cronName)
			if err != nil {
				core.LogErrorAndExit(logger, err)
			}
			cronRow.Cron.Name = cronName
			cronRow.Cron.ID = id
		}
		logger.Info("Registering run")

		dir, err := config.GetLogDir()
		if err != nil {
			panic(err)
		}
		stdout, err := os.CreateTemp(dir, "stdout.*.log")
		if err != nil {
			panic(err)
		}
		stdoutLog, err := os.OpenFile(stdout.Name(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}

		logger.Info("Run log created at: " + stdout.Name())
		runId, err := queries.StartRun(context.Background(), data.StartRunParams{
			CronID:      cronRow.Cron.ID,
			LogFile:     stdout.Name(),
			ExecLogFile: logFile,
		})
		if err != nil {
			panic(err)
		}

		logger.Info("Run created with ID " + strconv.FormatInt(runId, 10))

		// time.Sleep(10 * time.Second)
		cmdArgs := strings.Split(args[0], " ")
		runCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		runCmd.Stdout = stdoutLog
		runCmd.Stderr = stdoutLog
		err = runCmd.Run()
		status := core.RunStatusRunning
		if err != nil {
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
			ok, err := notify.NotifyRunSlack(slackToken, channel, completedRun)
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

	execCmd.Flags().String("cron-name", "", "Cron Name")
	if err := execCmd.MarkFlagRequired("cron-name"); err != nil {
		panic(err)
	}
	execCmd.Flags().Bool("skip-crontab", false, "Skips checking of crontab for registered CRON")
	execCmd.Flags().Bool("notify", false, "Notifies of the exec success")
}
