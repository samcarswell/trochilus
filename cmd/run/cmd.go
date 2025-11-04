/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"github.com/samcarswell/trochilus/cmd"
	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Commands related to crons",
}

func init() {
	cmd.RootCmd.AddCommand(RunCmd)
}
