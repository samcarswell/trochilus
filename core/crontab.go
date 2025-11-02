package core

import (
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

type CrontabItem struct {
	Name    string
	Command string
}

func GetCrontabItems(logger *slog.Logger) ([]CrontabItem, error) {
	runCmd := exec.Command("crontab", "-l")
	stdout, err := runCmd.Output()
	strOutput := string(stdout)
	if err != nil {
		return nil, fmt.Errorf("Unable to read crontab using: 'crontab -l': %w", err)
	}
	return parseCrontab(strOutput, logger)
}

func parseCrontab(crontab string, logger *slog.Logger) ([]CrontabItem, error) {
	cronRows := strings.Split(crontab, "\n")
	cronItems := []CrontabItem{}
	for _, value := range cronRows {
		cronRow, err := parseRow(value)
		if err != nil {
			logger.Error(err.Error())
			continue
		}
		if cronRow == nil {
			continue
		}
		cronItems = append(cronItems, *cronRow)
	}
	return cronItems, nil
}

func parseRow(row string) (*CrontabItem, error) {
	const cliName = "troc"
	const cronNameOptName = "--cron-name"
	rowStr := strings.TrimSpace(row)
	if len(rowStr) == 0 {
		return nil, nil
	}
	if strings.HasPrefix(rowStr, "#") {
		return nil, nil
	}
	if !strings.Contains(rowStr, cliName) {
		return nil, nil
	}

	fragments := strings.Split(rowStr, cronNameOptName)
	if len(fragments) < 2 {
		return nil, errors.New("Issue reading crontab line: " + rowStr)
	}
	name := strings.Split(strings.TrimSpace(fragments[1]), " ")[0]
	name = strings.ReplaceAll(name, "\"", "")
	name = strings.ReplaceAll(name, "'", "")
	return &CrontabItem{
		Name:    name,
		Command: rowStr,
	}, nil
}
