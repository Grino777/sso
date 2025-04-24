package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Grino777/sso/internal/app"
	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/lib/logger"
	storageU "github.com/Grino777/sso/internal/utils/storage"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	storageU.CheckStorageFolder()
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
