/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"carswellpress.com/cron-cowboy/cmd"
	"github.com/spf13/cobra"
)

// RunCmd represents the cron command
var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Commands related to crons",
}

func init() {
	cmd.RootCmd.AddCommand(RunCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// cronCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// cronCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
