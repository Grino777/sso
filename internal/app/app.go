package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/services/auth"
	"sso/internal/storage"
)

type App struct {
	Config     *config.Config
	Logger     *slog.Logger
	GRPCServer *grpcapp.App
	Storage    *storage.Storage
}

func New(
	cfg *config.Config,
	log *slog.Logger,
) *App {
	storage := storage.New("sqlite3", cfg.DB.Storage_path)
	defer storage.Close()

	authService := auth.New(log, storage, storage, storage, cfg.TokenTTL)

	gRPCServer := grpcapp.New(log, authService, int(cfg.GRPC.Port))

	return &App{
		Config:     cfg,
		Logger:     log,
		GRPCServer: gRPCServer,
		Storage:    storage,
	}
}
