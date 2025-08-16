/*
Copyright © 2025 Samuel Carswell <samuelrcarswell@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"carswellpress.com/cron-cowboy/cmd"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a CRON command",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(args)
		if len(args) == 0 {
			fmt.Println("Must have args")
			os.Exit(1)
		}
		runCmd := exec.Command(args[0], args[1:]...)
		stdout, err := runCmd.Output()
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		fmt.Println(string(stdout))
	},
}

func init() {
	cmd.RootCmd.AddCommand(runCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
