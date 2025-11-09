package opts

import (
	"errors"
	"log/slog"

	"github.com/samcarswell/trochilus/core"
	"github.com/spf13/cobra"
)

func GetStringOptOrExit(cmd *cobra.Command, name string) string {
	optVal, err := cmd.Flags().GetString(name)
	if err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
	return optVal
}
func GetBoolOptOrExit(cmd *cobra.Command, name string) bool {
	optVal, err := cmd.Flags().GetBool(name)
	if err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
	return optVal
}
func GetInt64OrExit(cmd *cobra.Command, name string) int64 {
	optVal, err := cmd.Flags().GetInt64(name)
	if err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
	if optVal == 0 {
		core.LogErrorAndExit(slog.Default(), errors.New("option "+name+" must have a value"))
	}
	return optVal
}
