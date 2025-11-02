package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"carswellpress.com/trochilus/config"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add cron",
	Run: func(cmd *cobra.Command, args []string) {
		cronName, err := cmd.Flags().GetString("name")
		if err != nil {
			panic("Could not get cron name")
		}
		queries := config.GetDatabase(cmd.Context())

		newCronId, err := queries.CreateCron(cmd.Context(), cronName)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: crons.name") {
				panic("Cron with name " + cronName + " already exists.")
			}
			panic(err)
		}

		fmt.Println("Cron created with ID " + strconv.FormatInt(newCronId, 10) + " and name " + cronName)
	},
}

func init() {
	CronCmd.AddCommand(addCmd)
	addCmd.Flags().String("name", "", "Cron Name")
	if err := addCmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
}
