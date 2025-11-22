package cmd

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/data"
	"github.com/samcarswell/trochilus/test"
)

func Test_execRunNonExistentJob(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	run := execRun(
		ctx,
		logger,
		jobName,
		false,
		conf,
		db,
		logFile,
		[]string{"./testdata/script-passes"},
	)
	dbJob, err := db.GetJob(ctx, jobName)
	if err != nil {
		t.Fatal(err.Error())
	}
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertString(t, jobName, run.Job.Name)
	test.AssertBool(t, false, run.Job.NotifyLogContent)
	test.AssertFileExists(t, run.Run.ExecLogFile)
	test.AssertFileExists(t, run.Run.LogFile)
	test.AssertString(t, "Succeeded", run.Run.Status)
	test.AssertFileContents(t, "Output line 1\nOutput line 2\n", run.Run.LogFile)
	test.AssertLogHasInfo(t, "Job not registered. Creating new job with name "+jobName, run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run log created at: "+run.Run.LogFile, run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run created with ID 1", run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run 1 completed: Succeeded", run.Run.ExecLogFile)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", run.Run.ExecLogFile)
	test.AssertFileInDirectory(t, conf.LogDir, run.Run.LogFile)
	lockRow := test.GetInfoLogLineStartingWith(t, "Created job lock at ", run.Run.ExecLogFile)
	lockFile := strings.Split(lockRow.Msg, "Created job lock at ")[1]
	test.AssertFileInDirectory(t, conf.LockDir, lockFile)

	test.AssertInt(t, 1, dbJob.Job.ID)
	test.AssertBool(t, false, dbJob.Job.NotifyLogContent)
	test.AssertString(t, jobName, dbJob.Job.Name)

	test.AssertInt(t, 1, dbRun.Run.ID)
	test.AssertInt(t, 1, dbRun.Run.JobID)
	test.AssertString(t, "Succeeded", dbRun.Run.Status)
	test.AssertString(t, run.Run.LogFile, dbRun.Run.LogFile)
	test.AssertString(t, run.Run.ExecLogFile, dbRun.Run.ExecLogFile)
}

func Test_execRunExistentJob(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}

	jobId, err := db.CreateJob(ctx, data.CreateJobParams{
		Name:             jobName,
		NotifyLogContent: false,
	})
	if err != nil {
		t.Fatal(err.Error())
	}

	run := execRun(
		ctx,
		logger,
		jobName,
		false,
		conf,
		db,
		logFile,
		[]string{"./testdata/script-passes"},
	)
	dbJob, err := db.GetJob(ctx, jobName)
	if err != nil {
		t.Fatal(err.Error())
	}
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertString(t, jobName, run.Job.Name)
	test.AssertInt(t, jobId, run.Job.ID)
	test.AssertBool(t, false, run.Job.NotifyLogContent)
	test.AssertFileExists(t, run.Run.ExecLogFile)
	test.AssertFileExists(t, run.Run.LogFile)
	test.AssertString(t, "Succeeded", run.Run.Status)
	test.AssertFileContents(t, "Output line 1\nOutput line 2\n", run.Run.LogFile)
	test.AssertLogDoesNotHaveInfo(t, "Job not registered. Creating new job with name "+jobName, run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run log created at: "+run.Run.LogFile, run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run created with ID 1", run.Run.ExecLogFile)
	test.AssertLogHasInfo(t, "Run 1 completed: Succeeded", run.Run.ExecLogFile)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", run.Run.ExecLogFile)

	test.AssertInt(t, 1, dbJob.Job.ID)
	test.AssertBool(t, false, dbJob.Job.NotifyLogContent)
	test.AssertString(t, jobName, dbJob.Job.Name)

	test.AssertInt(t, 1, dbRun.Run.ID)
	test.AssertInt(t, jobId, dbRun.Run.JobID)
	test.AssertString(t, "Succeeded", dbRun.Run.Status)
	test.AssertString(t, run.Run.LogFile, dbRun.Run.LogFile)
	test.AssertString(t, run.Run.ExecLogFile, dbRun.Run.ExecLogFile)
}

func Test_execRunScriptFails(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	run := execRun(
		ctx,
		logger,
		jobName,
		false,
		conf,
		db,
		logFile,
		[]string{"./testdata/script-fails"},
	)
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertString(t, "Failed", run.Run.Status)
	test.AssertString(t, "Failed", dbRun.Run.Status)
	test.AssertLogHasInfo(t, "Run 1 completed: Failed", run.Run.ExecLogFile)
	test.AssertFileContents(t, "This script will fail\n", run.Run.LogFile)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", run.Run.ExecLogFile)
}

func Test_execRunStdoutStderr(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	run := execRun(
		ctx,
		logger,
		jobName,
		false,
		conf,
		db,
		logFile,
		[]string{"./testdata/script-stdout-stderr"},
	)
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertString(t, "Succeeded", run.Run.Status)
	test.AssertString(t, "Succeeded", dbRun.Run.Status)
	test.AssertLogHasInfo(t, "Run 1 completed: Succeeded", run.Run.ExecLogFile)
	test.AssertFileContents(t, "Logging to stdout\nLogging to stderr\n", run.Run.LogFile)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", run.Run.ExecLogFile)
}

func Test_execRunSkippedRun(t *testing.T) {
	blocked := make(chan data.GetRunRow)
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile1, logger1 := test.CreateSysLogFile(t)
	logFile2, logger2 := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	go func() {
		blocked <- execRun(
			ctx,
			logger1,
			jobName,
			false,
			conf,
			db,
			logFile1,
			[]string{"./testdata/script-sleeps"},
		)
	}()
	time.Sleep(5 * time.Millisecond)
	skippedRun := execRun(
		ctx,
		logger2,
		jobName,
		true,
		conf,
		db,
		logFile2,
		[]string{"./testdata/script-passes"},
	)
	successfulRun := <-blocked
	runs, err := db.GetRuns(ctx, "")
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertInt(t, 2, len(runs))
	test.AssertString(t, "Skipped", skippedRun.Run.Status)
	test.AssertString(t, "", skippedRun.Run.LogFile)
	test.AssertLogHasWarn(t, "Skipping run 2. Job is already running.", skippedRun.Run.ExecLogFile)
	test.AssertString(t, "Succeeded", successfulRun.Run.Status)
	test.AssertLogHasInfo(t, "Run 1 completed: Succeeded", successfulRun.Run.ExecLogFile)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", skippedRun.Run.ExecLogFile)
	test.AssertLogDoesNotHaveInfo(t, "Sending notify message", successfulRun.Run.ExecLogFile)
}

func Test_execRunComplexCommand(t *testing.T) {
	ctx := context.Background()
	db := test.CreateDb(ctx, t)
	jobName := test.UniqueIdentifer()
	logFile, logger := test.CreateSysLogFile(t)
	conf := config.Config{
		LockDir: t.TempDir(),
		LogDir:  t.TempDir(),
	}
	run := execRun(
		ctx,
		logger,
		jobName,
		false,
		conf,
		db,
		logFile,
		[]string{"echo \"Testing again...\" && echo \"and again...\" | awk '{ print toupper($0) }'"},
	)
	dbRun, err := db.GetRun(ctx, 1)
	if err != nil {
		t.Fatal(err.Error())
	}

	test.AssertString(t, "Succeeded", run.Run.Status)
	test.AssertString(t, "Succeeded", dbRun.Run.Status)
	test.AssertLogHasInfo(t, "Run 1 completed: Succeeded", run.Run.ExecLogFile)
	test.AssertFileContents(t, "Testing again...\nAND AGAIN...\n", run.Run.LogFile)
}
