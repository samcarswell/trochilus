package cmd

import (
	"context"

	"carswellpress.com/trochilus/config"
	"github.com/rodaine/table"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Lists crons",
	Run: func(cmd *cobra.Command, args []string) {
		queries := config.GetDatabase(cmd.Context())

		cronRows, err := queries.GetCrons(context.Background())
		if err != nil {
			panic(err)
		}
		tbl := table.New("ID", "Name")
		for _, cron := range cronRows {
			tbl.AddRow(cron.Cron.ID, cron.Cron.Name)
		}
		tbl.Print()

	},
}

func init() {
	CronCmd.AddCommand(listCmd)
}
