package cmd

import (
	"context"
	"log"

	"github.com/rodaine/table"
	"github.com/samcarswell/trochilus/config"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists crons",
	Run: func(cmd *cobra.Command, args []string) {
		queries := config.GetDatabase(cmd.Context())

		cronRows, err := queries.GetCrons(context.Background())
		if err != nil {
			log.Fatalf("Unable to get crons %s", err)
		}
		tbl := table.New("ID", "Name", "Notify Log Content")
		for _, cron := range cronRows {
			tbl.AddRow(cron.Cron.ID, cron.Cron.Name, cron.Cron.NotifyLogContent)
		}
		tbl.Print()

	},
}

func init() {
	CronCmd.AddCommand(listCmd)
}
