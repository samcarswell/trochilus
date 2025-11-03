package cmd

import (
	"context"

	"carswellpress.com/trochilus/config"
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
			panic(err)
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
				run.Run.StartTime,
				run.Run.EndTime,
				run.Run.LogFile,
				run.Run.ExecLogFile,
				run.Run.Status,
			)
		}
		tbl.Print()

	},
}

func init() {
	RunCmd.AddCommand(listCmd)
}
