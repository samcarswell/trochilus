/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"carswellpress.com/cron-cowboy/cmd"
	"carswellpress.com/cron-cowboy/config"
	"carswellpress.com/cron-cowboy/data"
	"carswellpress.com/cron-cowboy/notify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO: should maybe have an optional flag on this to skip crontab lookup
// execCmd represents the run command
var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Run a CRON command",
	Run: func(cmd *cobra.Command, args []string) {
		cronName, err := cmd.Flags().GetString("name")
		if err != nil {
			panic("Could not get cron name")
		}
		skipCrontab, err := cmd.Flags().GetBool("skip-crontab")
		if err != nil {
			panic("Could not get skip-crontab flag")
		}

		logger, ok := config.LoggerFromContext(cmd.Context())
		if !ok {
			panic("Could not get logger from context")
		}
		logFile, ok := config.LogFileFromContext(cmd.Context())
		if !ok {
			panic("Could not get logFile from context")
		}

		if len(args) == 0 {
			logger.Error("Must provide args")
			os.Exit(1)
		}

		crons := getCrontab()

		// TODO: need to do a database lookup on the cron name here

		found := false
		for _, value := range crons {
			if value.Name == cronName {
				found = true
			}
		}
		if !found {
			// TODO: we might be able to continue here
			if !skipCrontab {
				panic("Unable to find cron defined with name " + cronName)
			}
		}

		queries := config.GetDatabase(cmd.Context())
		cronRow, err := queries.GetCron(context.Background(), cronName)
		if err != nil {
			if err == sql.ErrNoRows {
				logger.Info("Cron not registered")
			} else {
				panic(err)
			}
		}
		if cronRow == (data.GetCronRow{}) {
			logger.Info("Registering cron")
			id, err := queries.CreateCron(context.Background(), cronName)
			if err != nil {
				panic(err)
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
		stderr, err := os.CreateTemp(dir, "stderr.*.log")
		if err != nil {
			panic(err)
		}
		stdoutLog, err := os.OpenFile(stdout.Name(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}
		stderrLog, err := os.OpenFile(stderr.Name(), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			panic(err)
		}

		logger.Info("Stdout log created at: " + stdout.Name())
		logger.Info("Stderr log created at: " + stderr.Name())
		runId, err := queries.StartRun(context.Background(), data.StartRunParams{
			CronID:        cronRow.Cron.ID,
			StdoutLogFile: stdout.Name(),
			StderrLogFile: stderr.Name(),
			ExecLogFile:   logFile,
		})
		if err != nil {
			panic(err)
		}

		logger.Info("Run created with ID " + strconv.FormatInt(runId, 10))

		// time.Sleep(2 * time.Second)
		fmt.Println(args)
		cmdArgs := strings.Split(args[0], " ")
		runCmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		runCmd.Stdout = stdoutLog
		runCmd.Stderr = stderrLog
		err = runCmd.Run()
		succeeded := true
		if err != nil {
			succeeded = false
		}
		queries.EndRun(context.Background(), data.EndRunParams{
			Succeeded: succeeded,
			ID:        runId,
		})
		logger.Info("Run " + strconv.FormatInt(runId, 10) + " completed. Success: " + strconv.FormatBool(succeeded))

		slackToken := viper.GetString("slack.token")
		fmt.Println(slackToken)

		completedRun, err := queries.GetRun(cmd.Context(), runId)
		if err != nil {
			panic(err)
		}

		ok, err = notify.NotifyRunSlack(slackToken, "social", completedRun)
		if err != nil {
			panic(err)
		}
		if !ok {
			panic("Unable to notify slack")
		}
	},
}

type cronItem struct {
	Name string
}

func getCrontab() []cronItem {
	runCmd := exec.Command("crontab", "-l")
	stdout, err := runCmd.Output()
	strOutput := string(stdout)
	if err != nil {
		// TODO: this should actually check that the crontab does not exist
		fmt.Println(string(stdout))
		os.Exit(1)
	}
	return parseCrontab(strOutput)
}

// test for commented lines
func parseCrontab(crontab string) []cronItem {
	cronRows := strings.Split(crontab, "\n")
	cronItems := []cronItem{}
	for _, value := range cronRows {
		cronRow := parseRow(value)
		if cronRow == nil {
			continue
		}
		cronItems = append(cronItems, *cronRow)
	}
	return cronItems

}

func parseRow(value string) *cronItem {
	// TODO: this will break as soon as options are swapped
	const cronDef = "cron-cowboy exec --name"
	rowStr := strings.TrimSpace(value)
	if len(rowStr) == 0 {
		return nil
	}
	if strings.HasPrefix(rowStr, "#") {
		return nil
	}
	if !strings.Contains(rowStr, cronDef) {
		return nil
	}

	fragments := strings.Split(rowStr, cronDef)
	if len(fragments) < 2 {
		panic("TODO: Issue with crontab config")
	}
	name := strings.Split(strings.TrimSpace(fragments[1]), " ")[0]
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "'", "")
	return &cronItem{Name: name}
}

func init() {
	cmd.RootCmd.AddCommand(execCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	execCmd.Flags().String("name", "", "Cron Name")
	if err := execCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	execCmd.Flags().Bool("skip-crontab", false, "Skips checking of crontab for registered CRON")
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
