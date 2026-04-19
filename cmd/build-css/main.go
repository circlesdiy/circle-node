package main

import (
	"os"

	"circles.diy/internal/utils"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync() //nolint:errcheck
	if err := utils.BuildCSS(logger); err != nil {
		logger.Fatal("build css", zap.Error(err))
	}
}
