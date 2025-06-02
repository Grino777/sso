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
	log := logger.New(os.Stdout, slog.LevelDebug)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	errChan := make(chan error, 1)

	app, err := app.NewApp(log)
	if err != nil {
		log.Error("failed to creating app instance", logger.Error(err))
		os.Exit(1)
	}

	app.Run(errChan)

	select {
	case <-stop:
		log.Info("gracefully shutdown")
	case <-errChan:
		log.Error("stop app due to error")
	}

	app.Stop()
}
