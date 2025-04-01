package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
)

type App struct {
	Config     *config.Config
	Logger     *slog.Logger
	GRPCServer *grpcapp.App
}

func New(
	cfg *config.Config,
	log *slog.Logger,
) *App {
	gRPCServer := grpcapp.New(cfg, log)

	return &App{
		Config:     cfg,
		Logger:     log,
		GRPCServer: gRPCServer,
	}
}
