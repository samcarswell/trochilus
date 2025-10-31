package opts

import (
	"log/slog"

	"carswellpress.com/cron-cowboy/core"
	"github.com/spf13/cobra"
)

func GetStringOptOrExit(logger *slog.Logger, cmd *cobra.Command, name string) string {
	optVal, err := cmd.Flags().GetString(name)
	if err != nil {
		core.LogErrorAndExit(logger, err)
	}
	return optVal
}
func GetBoolOptOrExit(logger *slog.Logger, cmd *cobra.Command, name string) bool {
	optVal, err := cmd.Flags().GetBool(name)
	if err != nil {
		core.LogErrorAndExit(logger, err)
	}
	return optVal
}
