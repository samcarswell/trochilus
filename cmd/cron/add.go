package cmd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/data"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add cron",
	Run: func(cmd *cobra.Command, args []string) {
		cronName, err := cmd.Flags().GetString("name")
		if err != nil {
			log.Fatalf("Could not get cron name %s", err)
		}
		notifyLog, err := cmd.Flags().GetBool("notify-log")
		if err != nil {
			log.Fatalf("Could not get notify-log %s", err)
		}
		queries := config.GetDatabase(cmd.Context())

		newCronId, err := queries.CreateCron(cmd.Context(), data.CreateCronParams{
			Name:             cronName,
			NotifyLogContent: notifyLog,
		})
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: crons.name") {
				log.Fatalf("Cron with name "+cronName+" already exists. %s", err)
			}
			log.Fatalf("Unable to create cron. %s", err)
		}

		fmt.Println("Cron created with ID " + strconv.FormatInt(newCronId, 10) + " and name " + cronName)
	},
}

func init() {
	CronCmd.AddCommand(addCmd)
	addCmd.Flags().String("name", "", "Cron Name (required)")
	addCmd.Flags().Bool("notify-log", false, "Includes log output in notification message (default false)")
	if err := addCmd.MarkFlagRequired("name"); err != nil {
		log.Fatalf("Unable to mark name as required %s", err)
	}
}
