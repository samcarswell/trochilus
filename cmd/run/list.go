package cmd

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
	"github.com/samcarswell/trochilus/opts"
	"github.com/spf13/cobra"
)

var nameOpt = "name"
var statusField = "Status"

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
		t := core.NewTable()
		t.AppendHeader(table.Row{
			"ID",
			"Job Name",
			"Start Time",
			"End Time",
			"Log File",
			"Exec Log File",
			statusField,
		})
		for _, run := range runRows {
			t.AppendRow(table.Row{
				run.Run.ID,
				run.Job.Name,
				core.FormatTime(run.Run.StartTime, conf.LocalTime),
				core.FormatTime(run.Run.EndTime.Time, conf.LocalTime),
				run.Run.LogFile,
				run.Run.ExecLogFile,
				core.FormatStatus(core.RunStatus(run.Run.Status), conf.Display.Emoji),
			})
		}

		statusTransformer := text.Transformer(func(val interface{}) string {
			if status, ok := val.(string); ok {
				color := text.FgWhite
				switch status {
				case core.FormatStatus(core.RunStatusSucceeded, conf.Display.Emoji):
					if conf.Display.Color.Status.Succeeded {
						color = text.FgGreen
					}
				case core.FormatStatus(core.RunStatusFailed, conf.Display.Emoji):
					if conf.Display.Color.Status.Failed {
						color = text.FgHiRed
					}
				case core.FormatStatus(core.RunStatusRunning, conf.Display.Emoji):
					if conf.Display.Color.Status.Running {
						color = text.FgCyan
					}
				case core.FormatStatus(core.RunStatusSkipped, conf.Display.Emoji):
					if conf.Display.Color.Status.Skipped {
						color = text.FgYellow
					}
				case core.FormatStatus(core.RunStatusTerminated, conf.Display.Emoji):
					if conf.Display.Color.Status.Terminated {
						color = text.FgHiMagenta
					}
				}
				return color.Sprintf("%s", status)
			}
			return fmt.Sprint(val)
		})

		t.SetColumnConfigs([]table.ColumnConfig{
			{
				Name:        statusField,
				Transformer: statusTransformer,
			},
		})

		t.Render()

	},
}

func init() {
	RunCmd.AddCommand(listCmd)

	listCmd.Flags().String(nameOpt, "", "Name of job to filter on")
}
