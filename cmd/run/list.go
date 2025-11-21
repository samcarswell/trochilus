package cmd

import (
	"database/sql"
	"errors"
	"log/slog"

	"github.com/rodaine/table"
	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
	"github.com/samcarswell/trochilus/opts"
	"github.com/spf13/cobra"
)

var nameOpt = "name"

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists runs",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		cronName := opts.GetStringOptOrExit(cmd, nameOpt)
		queries := config.GetDatabase(cmd.Context())

		_, err := queries.GetCron(cmd.Context(), cronName)
		if err != nil {
			if err == sql.ErrNoRows {
				core.LogErrorAndExit(logger, errors.New("cron with name '"+cronName+"' not found"))
			} else {
				core.LogErrorAndExit(logger, err)
			}
		}

		runRows, err := queries.GetRuns(cmd.Context(), cronName)
		if err != nil {
			core.LogErrorAndExit(slog.Default(), err)
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

	listCmd.Flags().String(nameOpt, "", "Name of cron to filter on")
}
