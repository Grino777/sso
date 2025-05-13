package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/Grino777/sso/internal/app"
	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/lib/logger"
	appUtils "github.com/Grino777/sso/internal/utils/app"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	log := logger.New(os.Stdout, slog.LevelDebug)
	cfg := config.Load(log)

	if err := appUtils.CheckKeysFolder(cfg.KeysDir); err != nil {
		panic(err.Error())
	}

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
