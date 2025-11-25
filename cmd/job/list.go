package cmd

import (
	"context"
	"errors"
	"log/slog"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List jobs",
	Run: func(cmd *cobra.Command, args []string) {
		queries := config.GetDatabase(cmd.Context())

		jobRows, err := queries.GetJobs(context.Background())
		if err != nil {
			core.LogErrorAndExit(slog.Default(), err, errors.New("unable to get jobs"))
		}
		t := core.NewTable()
		t.AppendHeader(table.Row{"ID", "Name", "Notify Log Content"})
		for _, job := range jobRows {
			t.AppendRow(table.Row{job.Job.ID, job.Job.Name, job.Job.NotifyLogContent})
		}
		t.Render()
	},
}

func init() {
	JobCmd.AddCommand(listCmd)
}
