package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/samcarswell/trochilus/config"
	"github.com/samcarswell/trochilus/core"
	"github.com/samcarswell/trochilus/opts"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watches a run",
	Long:  "Watches the logs of a run. If is not running, the log will be printed and the command will immediately exit",
	Run: func(cmd *cobra.Command, args []string) {
		logger := slog.Default()
		queries := config.GetDatabase(cmd.Context())
		runId := opts.GetInt64OrExit(cmd, "run-id")
		runRow, err := queries.GetRun(cmd.Context(), runId)

		file, err := os.Open(runRow.Run.LogFile)
		if err != nil {
			return
		}
		defer file.Close()
		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					// without this sleep you would hogg the CPU
					time.Sleep(500 * time.Millisecond)
					// truncated ?
					truncated, errTruncated := isTruncated(file)
					if errTruncated != nil {
						break
					}
					if truncated {
						// seek from start
						_, errSeekStart := file.Seek(0, io.SeekStart)
						if errSeekStart != nil {
							break
						}
					}
					finished, err := queries.IsRunFinished(cmd.Context(), runId)
					if err != nil {
						logger.Error(err.Error())
					}
					if finished {
						logger.Info("Run has finished. Exiting.")
						break
					}
					continue
				}
				break
			}
			fmt.Printf("%s", string(line))
		}
	},
}

func init() {
	watchCmd.Flags().Int64P("run-id", "r", 0, "Run Id")
	if err := watchCmd.MarkFlagRequired("run-id"); err != nil {
		core.LogErrorAndExit(slog.Default(), err)
	}
	RunCmd.AddCommand(watchCmd)
}

// https://medium.com/@arunprabhu.1/tailing-a-file-in-golang-72944204f22b
func isTruncated(file *os.File) (bool, error) {
	// current read position in a file
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}
	// file stat to get the size
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	return currentPos > fileInfo.Size(), nil
}
