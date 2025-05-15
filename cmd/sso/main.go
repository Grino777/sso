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

	app, err := app.New(log)
	if err != nil {
		panic(err.Error())
	}

	go func() {
		app.GRPCServer.MustRun()
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	app.GRPCServer.Stop()
	log.Info("Gracefully stopped")
}
