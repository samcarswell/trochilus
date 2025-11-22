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
		conf := config.GetConfig()
		jobName := opts.GetStringOptOrExit(cmd, nameOpt)
		queries := config.GetDatabase(cmd.Context())

		if jobName != "" {
			_, err := queries.GetJob(cmd.Context(), jobName)
			if err != nil {
				if err == sql.ErrNoRows {
					core.LogErrorAndExit(logger, errors.New("job with name '"+jobName+"' not found"))
				} else {
					core.LogErrorAndExit(logger, err)
				}
			}
		}

		runRows, err := queries.GetRuns(cmd.Context(), jobName)
		if err != nil {
			core.LogErrorAndExit(slog.Default(), err)
		}
		tbl := table.New(
			"ID",
			"Job Name",
			"Start Time",
			"End Time",
			"Log File",
			"Exec Log File",
			"Status",
		)
		for _, run := range runRows {
			tbl.AddRow(
				run.Run.ID,
				run.Job.Name,
				core.FormatTime(run.Run.StartTime, conf.LocalTime),
				core.FormatTime(run.Run.EndTime.Time, conf.LocalTime),
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

	listCmd.Flags().String(nameOpt, "", "Name of job to filter on")
}
