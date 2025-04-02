package main

import (
	"database/sql"
	"log/slog"
	"os"
	"os/signal"
	"sso/internal/app"
	"sso/internal/config"
	"sso/internal/logger"
	"sso/migrations"
	"syscall"

	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.Load()

	db, err := sql.Open("sqlite3", cfg.DB.Storage_path)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	defer db.Close()

	migrate(db)

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

// Performs migrations
func migrate(db *sql.DB) {

	embedMigrations := migrations.GetEmbedFS()

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		panic("failed to set dialect: " + err.Error())
	}

	if err := goose.Up(db, "migrations"); err != nil {
		panic("failed to apply migrations: " + err.Error())
	}
}
