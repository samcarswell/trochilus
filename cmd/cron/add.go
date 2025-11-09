package cmd

import (
	"errors"
	"log"
	"log/slog"
	"strconv"
	"strings"

	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
	"github.com/samcarswell/trochilus/data"
	"github.com/samcarswell/trochilus/opts"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add cron",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		cronName := opts.GetStringOptOrExit(cmd, "name")
		notifyLog := opts.GetBoolOptOrExit(cmd, "notify-log")
		queries := config.GetDatabase(cmd.Context())

		newCronId, err := queries.CreateCron(cmd.Context(), data.CreateCronParams{
			Name:             cronName,
			NotifyLogContent: notifyLog,
		})
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed: crons.name") {
				core.LogErrorAndExit(logger, err, errors.New("cron with name "+cronName+" already exists"))
			}
			core.LogErrorAndExit(logger, err, errors.New("unable to create cron"))
		}

		logger.Info("Cron created with ID " + strconv.FormatInt(newCronId, 10) + " and name " + cronName)
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
