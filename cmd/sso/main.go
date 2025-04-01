package main

import (
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/logger"
	"syscall"
)

func main() {
	cfg := config.Load()
	log := logger.New(os.Stdout, slog.LevelDebug)

	app := app.New(cfg, log)

	go func() {
		app.GRPCServer.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	app.GRPCServer.Stop()
	log.Info("Gracefully stopped")
}
