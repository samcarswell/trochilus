package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/opts"
	"github.com/spf13/cobra"
)

type RunShow struct {
	ID            int64
	CronName      string
	StartTime     time.Time
	EndTime       time.Time
	LogFile       string
	SystemLogFile string
	Status        string
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show details of a run",
	Run: func(cmd *cobra.Command, args []string) {
		logger := config.GetLoggerOrExit(cmd.Context())
		runId := opts.GetInt64OrExit(logger, cmd, "run-id")
		queries := config.GetDatabase(cmd.Context())

		runRow, err := queries.GetRun(context.Background(), runId)
		if err != nil {
			log.Fatalf("No rows found: %s", err)
		}
		data := RunShow{
			ID:            runRow.Run.ID,
			CronName:      runRow.Cron.Name,
			StartTime:     runRow.Run.StartTime,
			EndTime:       runRow.Run.EndTime.Time,
			LogFile:       runRow.Run.LogFile,
			SystemLogFile: runRow.Run.ExecLogFile,
			Status:        runRow.Run.Status,
		}
		jsonData, err := json.MarshalIndent(data, "", "    ")
		if err != nil {
			log.Fatalln(err)
		}
		fmt.Println(string(jsonData))
	},
}

func init() {
	RunCmd.AddCommand(showCmd)

	showCmd.Flags().Int64P("run-id", "r", 0, "Run id")
	if err := showCmd.MarkFlagRequired("run-id"); err != nil {
		log.Fatal(err)
	}
}
