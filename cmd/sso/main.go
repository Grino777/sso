package main

import (
	"log/slog"
	"os"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/logger"
)

func main() {
	cfg := config.Load()
	logger := logger.New(os.Stdout, slog.LevelDebug)

	app := app.New(cfg, logger)
	// app.Run()

}
