package app

import (
	"sso/internal/config"
	"sso/internal/logger"
)

type App struct {
	Config *config.Config
	Logger *logger.Logger
	// GRPCServer *GRPCServer
}

func New(cfg *config.Config, logger *logger.Logger) *App {
	return &App{
		Config: cfg,
		Logger: logger,
	}
}

func (a *App) Run() error {
	// logger
	return nil
}
