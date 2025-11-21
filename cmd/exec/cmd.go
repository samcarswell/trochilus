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
	"os/signal"
	"path/filepath"
	"strconv"

	"github.com/gofrs/flock"
	"github.com/samcarswell/trochilus/cmd"
	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
	"github.com/samcarswell/trochilus/data"
	"github.com/samcarswell/trochilus/notify"
	"github.com/samcarswell/trochilus/opts"
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
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		cronName := opts.GetStringOptOrExit(cmd, nameOpt)
		notifyOpt := opts.GetBoolOptOrExit(cmd, notifyOpt)
		conf := config.GetConfig()
		queries := config.GetDatabase(cmd.Context())
		if notifyOpt && conf.Notify.Slack.Token == "" {
			core.LogErrorAndExit(logger, errors.New("notify is set but notify.slack.token is blank."))
		}
		if notifyOpt && conf.Notify.Slack.Channel == "" {
			core.LogErrorAndExit(logger, errors.New("notify is set but notify.slack.channel is blank."))
		}

		logFile := config.GetLogFileOrExit(logger, cmd.Context())

		if len(args) == 0 {
			core.LogErrorAndExit(logger, errors.New("must provide args"))
		}
		completedRun := execRun(
			cmd.Context(),
			logger,
			cronName,
			notifyOpt,
			conf,
			queries,
			logFile,
			args,
		)
		core.PrintJson(completedRun.Run)
	},
}

func init() {
	cmd.RootCmd.AddCommand(execCmd)

	execCmd.Flags().String(nameOpt, "", "Name of cron to execute. Will create it if it does not exist")
	if err := execCmd.MarkFlagRequired(nameOpt); err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
	execCmd.Flags().Bool(notifyOpt, false, "Notifies of the exec success")
}

func execRun(
	ctx context.Context,
	logger *slog.Logger,
	cronName string,
	isNotify bool,
	conf config.Config,
	db *data.Queries,
	logFile string,
	args []string,
) data.GetRunRow {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	cronRow, err := db.GetCron(ctx, cronName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Print("Cron not registered. Creating new Cron with name " + cronName)
		} else {
			core.LogErrorAndExit(logger, err)
		}
	}
	if cronRow == (data.GetCronRow{}) {
		id, err := db.CreateCron(context.Background(), data.CreateCronParams{
			Name:             cronName,
			NotifyLogContent: false,
		})
		if err != nil {
			core.LogErrorAndExit(logger, err)
		}
		cronRow.Cron.Name = cronName
		cronRow.Cron.ID = id
	}

	lockFile := filepath.Join(conf.LockDir, cronName+".lock")
	f := flock.New(lockFile)

	locked, err := f.TryLock()

	if err != nil || !locked {
		return skipRun(
			cronRow.Cron,
			logFile,
			conf.Notify,
			db,
			context.Background(),
			logger,
		)
	}
	if !locked {
		core.LogErrorAndExit(logger, errors.New("unable to create lock for cron. Likely already running"))
	}

	defer f.Unlock()
	logger.Info("Created cron lock at " + lockFile)

	stdout, err := os.CreateTemp(conf.LogDir, cronName+".*.log")
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to create log file"))
	}
	stdoutLog, err := os.OpenFile(stdout.Name(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to open log file"))
	}

	logger.Info("Run log created at: " + stdout.Name())
	runId, err := db.StartRun(context.Background(), data.StartRunParams{
		CronID:      cronRow.Cron.ID,
		LogFile:     stdout.Name(),
		ExecLogFile: logFile,
	})
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to start run"))
	}

	logger.Info("Run created with ID " + strconv.FormatInt(runId, 10))

	cmdArgs := []string{"-c"}
	cmdArgs = append(cmdArgs, args[0])
	runCmd := exec.Command("/bin/sh", cmdArgs...)
	runCmd.Stdout = stdoutLog
	runCmd.Stderr = stdoutLog
	err = runCmd.Run()
	status := core.RunStatusSucceeded
	if err != nil {
		logger.Error("Error occurred during run", "error", err)
		status = core.RunStatusFailed
	}
	db.EndRun(context.Background(), data.EndRunParams{
		Status: string(status),
		ID:     runId,
	})
	logger.Info("Run " + strconv.FormatInt(runId, 10) + " completed: " + string(status))

	completedRun, err := db.GetRun(ctx, runId)
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to get completed run"))
	}

	if isNotify {
		logger.Info("Sending notify message")
		ok, err := notify.NotifyRun(
			conf.Notify,
			notify.RunNotifyInfo{
				Name:             completedRun.Cron.Name,
				Id:               completedRun.Run.ID,
				Status:           core.RunStatus(completedRun.Run.Status),
				LogFile:          completedRun.Run.LogFile,
				NotifyLogContent: cronRow.Cron.NotifyLogContent,
			},
		)
		if err != nil {
			core.LogErrorAndExit(logger, err, errors.New("unable to notify"))
		}
		if !ok {
			logger.Error("command was run, but notification was unable to be sent")
		}
	}
	return completedRun
}

func skipRun(
	cron data.Cron,
	execLogFile string,
	notifyConf config.NotifyConfig,
	queries *data.Queries,
	ctx context.Context,
	logger *slog.Logger,
) data.GetRunRow {
	id, err := queries.SkipRun(ctx, data.SkipRunParams{
		CronID:      cron.ID,
		ExecLogFile: execLogFile,
	})
	if err != nil {
		// TODO: need a standard function here to deal with errors and communicate to slack
		core.LogErrorAndExit(logger, err, errors.New("unable to skip run"))
	}
	run, err := queries.GetRun(ctx, id)
	if err != nil {
		// TODO: need a standard function here to deal with errors and communicate to slack
		core.LogErrorAndExit(logger, err, errors.New("unable to get updated run"))
	}
	logger.Warn("Skipping run " + strconv.FormatInt(id, 10) + ". Cron is already running.")
	notify.NotifyRun(
		notifyConf,
		notify.RunNotifyInfo{
			Name:             run.Cron.Name,
			Id:               run.Run.ID,
			Status:           core.RunStatus(run.Run.Status),
			LogFile:          "",
			NotifyLogContent: cron.NotifyLogContent,
		},
	)
	row, err := queries.GetRun(ctx, run.Run.ID)
	if err != nil {
		core.LogErrorAndExit(logger, err, errors.New("unable to get updated run"))
	}
	return row
}
