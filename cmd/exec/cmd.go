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
	Run: func(cmd *cobra.Command, args []string) {
		logger := config.GetLoggerOrExit(cmd.Context())
		cronName := opts.GetStringOptOrExit(logger, cmd, nameOpt)
		notifyOpt := opts.GetBoolOptOrExit(logger, cmd, notifyOpt)
		notifyConf := config.GetNotifyConfig()
		queries := config.GetDatabase(cmd.Context())
		if notifyOpt && notifyConf.Slack.Token == "" {
			core.LogErrorAndExit(logger, errors.New("notify is set but slack.token is blank."))
		}
		if notifyOpt && notifyConf.Slack.Channel == "" {
			core.LogErrorAndExit(logger, errors.New("notify is set but slack.channel is blank."))
		}

		logFile := config.GetLogFileOrExit(logger, cmd.Context())

		if len(args) == 0 {
			core.LogErrorAndExit(logger, errors.New("Must provide args"))
		}
		logDir, err := config.GetLogDir()
		if err != nil {
			log.Fatalf("Unable to get logdir %s", err)
		}
		completedRun := execRun(
			cmd.Context(),
			logger,
			cronName,
			notifyOpt,
			notifyConf,
			queries,
			logFile,
			args,
			logDir,
		)
		core.PrintJson(completedRun.Run)
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

func execRun(
	ctx context.Context,
	logger *slog.Logger,
	cronName string,
	isNotify bool,
	notifyConf config.NotifyConfig,
	db *data.Queries,
	logFile string,
	args []string,
	logDir string,
) data.GetRunRow {
	cronRow, err := db.GetCron(ctx, cronName)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info("Cron not registered. Creating new Cron with name " + cronName)
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

	// TODO: lock dir should be configurable
	lockFile := filepath.Join(os.TempDir(), cronName+".lock")
	f := flock.New(lockFile)

	locked, err := f.TryLock()

	if err != nil || !locked {
		return skipRun(cronRow.Cron, logFile, notifyConf, db, context.Background(), logger)
	}
	if !locked {
		log.Fatalf("Unable to create lock for cron. Likely already running")
	}

	defer f.Unlock()
	logger.Info("Created cron lock at " + lockFile)

	stdout, err := os.CreateTemp(logDir, cronName+".*.log")
	if err != nil {
		log.Fatalf("Unable to create log file %s", err)
	}
	stdoutLog, err := os.OpenFile(stdout.Name(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		log.Fatalf("Unable to open log file %s", err)
	}

	logger.Info("Run log created at: " + stdout.Name())
	runId, err := db.StartRun(context.Background(), data.StartRunParams{
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
	db.EndRun(context.Background(), data.EndRunParams{
		Status: string(status),
		ID:     runId,
	})
	logger.Info("Run " + strconv.FormatInt(runId, 10) + " completed: " + string(status))

	completedRun, err := db.GetRun(ctx, runId)
	if err != nil {
		log.Fatalf("Unable to get completed run %s", err)
	}

	if isNotify {
		ok, err := notify.NotifyRun(
			notifyConf,
			notify.RunNotifyInfo{
				Name:             completedRun.Cron.Name,
				Id:               completedRun.Run.ID,
				Status:           core.RunStatus(completedRun.Run.Status),
				LogFile:          completedRun.Run.LogFile,
				NotifyLogContent: cronRow.Cron.NotifyLogContent,
			},
		)
		if err != nil {
			log.Fatalf("Unable to notify slack %s", err)
		}
		if !ok {
			logger.Error("Command was run, but slack was unable to be notified")
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
		log.Fatalf("Unable to skip run %s", err)
	}
	run, err := queries.GetRun(ctx, id)
	if err != nil {
		// TODO: need a standard function here to deal with errors and communicate to slack
		log.Fatalf("Unable to get updated run %s", err)
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
		log.Fatalf("Unable to query row %s", err)
	}
	return row
}
