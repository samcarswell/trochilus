package cmd

import (
	"context"
	"log"

	"carswellpress.com/trochilus/config"
	"carswellpress.com/trochilus/core"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists runs",
	Run: func(cmd *cobra.Command, args []string) {
		queries := config.GetDatabase(cmd.Context())

		runRows, err := queries.GetRuns(context.Background())
		if err != nil {
			log.Fatalf("Unable to find any runs %s", err)
		}
		tbl := table.New(
			"ID",
			"Cron Name",
			"Start Time",
			"End Time",
			"Log File",
			"Exec Log File",
			"Status",
		)
		for _, run := range runRows {
			tbl.AddRow(
				run.Run.ID,
				run.Cron.Name,
				core.FormatTime(run.Run.StartTime),
				core.FormatTime(run.Run.EndTime.Time),
				run.Run.LogFile,
				run.Run.ExecLogFile,
				core.FormatStatus(core.RunStatus(run.Run.Status)),
			)
		}
		tbl.Print()

	},
}

func init() {
	RunCmd.AddCommand(listCmd)
}
