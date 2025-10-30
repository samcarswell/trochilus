package core

import (
	"log/slog"
	"os"
)

func LogErrorAndExit(logger *slog.Logger, err error) {
	logger.Error(err.Error())
	os.Exit(1)
}
