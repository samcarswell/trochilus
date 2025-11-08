package cmd

import (
	"context"
	"testing"
	"time"

	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/data"
	"github.com/samcarswell/trochilus/test"
)

func Test_execRunNonExistentCron(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	cronName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	run := execRun(
		ctx,
		logger,
		cronName,
		false,
		config.NotifyConfig{},
		db,
		logFile,
		[]string{"./testdata/script-passes"},
		t.TempDir(),
	)
	dbCron, err := db.GetCron(ctx, cronName)
	if err != nil {
		t.Fatal(err.Error())
	}
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertString(t, cronName, run.Cron.Name)
	test.AssertBool(t, false, run.Cron.NotifyLogContent)
	test.AssertFileExists(t, run.Run.ExecLogFile)
	test.AssertFileExists(t, run.Run.LogFile)
	test.AssertString(t, "Succeeded", run.Run.Status)
	test.AssertFileContents(t, "Output line 1\nOutput line 2\n", run.Run.LogFile)
	test.AssertLogHasInfo(t, "Cron not registered. Creating new Cron with name "+cronName, run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run log created at: "+run.Run.LogFile, run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run created with ID 1", run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run 1 completed: Succeeded", run.Run.ExecLogFile)

	test.AssertInt(t, 1, dbCron.Cron.ID)
	test.AssertBool(t, false, dbCron.Cron.NotifyLogContent)
	test.AssertString(t, cronName, dbCron.Cron.Name)

	test.AssertInt(t, 1, dbRun.Run.ID)
	test.AssertInt(t, 1, dbRun.Run.CronID)
	test.AssertString(t, "Succeeded", dbRun.Run.Status)
	test.AssertString(t, run.Run.LogFile, dbRun.Run.LogFile)
	test.AssertString(t, run.Run.ExecLogFile, dbRun.Run.ExecLogFile)
}

func Test_execRunExistentCron(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	cronName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)

	cronId, err := db.CreateCron(ctx, data.CreateCronParams{
		Name:             cronName,
		NotifyLogContent: false,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	run := execRun(
		ctx,
		logger,
		cronName,
		false,
		config.NotifyConfig{},
		db,
		logFile,
		[]string{"./testdata/script-passes"},
		t.TempDir(),
	)
	dbCron, err := db.GetCron(ctx, cronName)
	if err != nil {
		t.Fatal(err.Error())
	}
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertString(t, cronName, run.Cron.Name)
	test.AssertInt(t, cronId, run.Cron.ID)
	test.AssertBool(t, false, run.Cron.NotifyLogContent)
	test.AssertFileExists(t, run.Run.ExecLogFile)
	test.AssertFileExists(t, run.Run.LogFile)
	test.AssertString(t, "Succeeded", run.Run.Status)
	test.AssertFileContents(t, "Output line 1\nOutput line 2\n", run.Run.LogFile)
	test.AssertLogDoesNotHaveInfo(t, "Cron not registered. Creating new Cron with name "+cronName, run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run log created at: "+run.Run.LogFile, run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run created with ID 1", run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run 1 completed: Succeeded", run.Run.ExecLogFile)

	test.AssertInt(t, 1, dbCron.Cron.ID)
	test.AssertBool(t, false, dbCron.Cron.NotifyLogContent)
	test.AssertString(t, cronName, dbCron.Cron.Name)

	test.AssertInt(t, 1, dbRun.Run.ID)
	test.AssertInt(t, cronId, dbRun.Run.CronID)
	test.AssertString(t, "Succeeded", dbRun.Run.Status)
	test.AssertString(t, run.Run.LogFile, dbRun.Run.LogFile)
	test.AssertString(t, run.Run.ExecLogFile, dbRun.Run.ExecLogFile)
}

func Test_execRunScriptFails(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	cronName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	run := execRun(
		ctx,
		logger,
		cronName,
		false,
		config.NotifyConfig{},
		db,
		logFile,
		[]string{"./testdata/script-fails"},
		t.TempDir(),
	)
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertString(t, "Failed", run.Run.Status)
	test.AssertString(t, "Failed", dbRun.Run.Status)
	test.AssertLogHasInfo(t, "Run 1 completed: Failed", run.Run.ExecLogFile)
	test.AssertFileContents(t, "This script will fail\n", run.Run.LogFile)
}

func Test_execRunStdoutStderr(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	cronName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	run := execRun(
		ctx,
		logger,
		cronName,
		false,
		config.NotifyConfig{},
		db,
		logFile,
		[]string{"./testdata/script-stdout-stderr"},
		t.TempDir(),
	)
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertString(t, "Succeeded", run.Run.Status)
	test.AssertString(t, "Succeeded", dbRun.Run.Status)
	test.AssertLogHasInfo(t, "Run 1 completed: Succeeded", run.Run.ExecLogFile)
	test.AssertFileContents(t, "Logging to stdout\nLogging to stderr\n", run.Run.LogFile)
}

func Test_execRunSkippedRun(t *testing.T) {
	blocked := make(chan data.GetRunRow)
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	cronName := test.UniqueIdentifer()
	logFile1, logger1 := test.CreateSysLogFile(t)
	logFile2, logger2 := test.CreateSysLogFile(t)
	go func() {
		blocked <- execRun(
			ctx,
			logger1,
			cronName,
			false,
			config.NotifyConfig{},
			db,
			logFile1,
			[]string{"./testdata/script-sleeps"},
			t.TempDir(),
		)
	}()
	time.Sleep(5 * time.Millisecond)
	skippedRun := execRun(
		ctx,
		logger2,
		cronName,
		false,
		config.NotifyConfig{},
		db,
		logFile2,
		[]string{"./testdata/script-passes"},
		t.TempDir(),
	)
	successfulRun := <-blocked
	runs, err := db.GetRuns(ctx)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertInt(t, 2, len(runs))
	test.AssertString(t, "Skipped", skippedRun.Run.Status)
	test.AssertString(t, "", skippedRun.Run.LogFile)
	test.AssertLogHasWarn(t, "Skipping run 2. Cron is already running.", skippedRun.Run.ExecLogFile)
	test.AssertString(t, "Succeeded", successfulRun.Run.Status)
	test.AssertLogHasInfo(t, "Run 1 completed: Succeeded", successfulRun.Run.ExecLogFile)
}
