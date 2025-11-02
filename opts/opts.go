package opts

import (
	"errors"
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
func GetInt64OrExit(logger *slog.Logger, cmd *cobra.Command, name string) int64 {
	optVal, err := cmd.Flags().GetInt64(name)
	if err != nil {
		core.LogErrorAndExit(logger, err)
	}
	if optVal == 0 {
		core.LogErrorAndExit(logger, errors.New("Option "+name+" must have a value"))
	}
	return optVal
}
