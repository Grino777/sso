package main

import (
	"context"
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
	ctx, cancel := context.WithCancel(context.Background())

	app, err := app.NewApp(ctx, log, cancel)
	if err != nil {
		log.Error("failed to creating app obj", logger.Error(err))
		os.Exit(1)
	}

	app.Run(ctx)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	app.Stop()
	log.Info("Gracefully stopped")
}
