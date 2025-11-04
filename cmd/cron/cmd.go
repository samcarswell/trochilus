/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"github.com/samcarswell/trochilus/cmd"
	"github.com/spf13/cobra"
)

var CronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Commands related to crons",
}

func init() {
	cmd.RootCmd.AddCommand(CronCmd)
}
