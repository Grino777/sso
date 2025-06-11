package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Grino777/sso/internal/app"
	"github.com/Grino777/sso/internal/lib/logger"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log := logger.NewLogger(os.Stdout, slog.LevelDebug)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	errChan := make(chan error, 5)

	app, err := app.NewApp(log)
	if err != nil {
		log.Error("failed to creating app instance", logger.Error(err))
		os.Exit(1)
	}

	go func() {
		app.Run(errChan)
	}()

	select {
	case <-stop:
		log.Info("gracefully shutting down")
		app.Stop()
	case err := <-errChan:
		log.Error("stopping app due to error", logger.Error(err))
		app.Stop()
	}
}
