package cmd

import (
	"database/sql"
	"errors"
	"log"
	"log/slog"

	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
	"github.com/samcarswell/trochilus/data"
	"github.com/samcarswell/trochilus/opts"
	"github.com/spf13/cobra"
)

var notifyLogOpt = "notify-log"
var newNameOpt = "new-name"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update cron",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		cronName := opts.GetStringOptOrExit(cmd, "name")
		queries := config.GetDatabase(cmd.Context())

		cron, err := queries.GetCron(cmd.Context(), cronName)
		if err != nil {
			if err == sql.ErrNoRows {
				core.LogErrorAndExit(logger, errors.New("cron with name '"+cronName+"' not found"))
			} else {
				core.LogErrorAndExit(logger, err)
			}
		}

		if cmd.Flags().Changed(notifyLogOpt) {
			cron.Cron.NotifyLogContent = opts.GetBoolOptOrExit(cmd, notifyLogOpt)
		}
		if cmd.Flags().Changed(newNameOpt) {
			cron.Cron.Name = opts.GetStringOptOrExit(cmd, newNameOpt)
		}

		err = queries.UpdateCron(cmd.Context(), data.UpdateCronParams{
			ID:               cron.Cron.ID,
			Name:             cron.Cron.Name,
			NotifyLogContent: cron.Cron.NotifyLogContent,
		})

		if err != nil {
			core.LogErrorAndExit(logger, err, errors.New("unable to update cron"))
		}

		logger.Info("Cron updated")
	},
}

func init() {
	CronCmd.AddCommand(updateCmd)
	updateCmd.Flags().String("name", "", "Cron Name (required)")
	updateCmd.Flags().String(newNameOpt, "", "New cron Name")
	updateCmd.Flags().Bool(notifyLogOpt, false, "Includes the raw log output rather than the log filename in notification messages (default false)")
	if err := updateCmd.MarkFlagRequired("name"); err != nil {
		log.Fatalf("Unable to mark name as required %s", err)
	}
}
