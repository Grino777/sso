package app

import (
	"log/slog"
	grpcapp "sso/internal/app/grpc"
	"sso/internal/config"
	"sso/internal/services/auth"
	reidsstore "sso/internal/storage/redis"
	sqlitestore "sso/internal/storage/sqlite"
)

type App struct {
	Config       *config.Config
	Logger       *slog.Logger
	GRPCServer   *grpcapp.App
	DBStorage    *sqlitestore.Storage
	RedisStorage *reidsstore.RedisStore
}

func New(
	cfg *config.Config,
	log *slog.Logger,
) *App {
	sqliteStore := sqlitestore.New("sqlite3", cfg.DB.Storage_path)
	reidsStore := reidsstore.New(cfg.Redis)

	authService := auth.New(log, sqliteStore, sqliteStore, sqliteStore, cfg.TokenTTL)

	gRPCServer := grpcapp.New(log, authService, sqliteStore, int(cfg.GRPC.Port))

	return &App{
		Config:       cfg,
		Logger:       log,
		GRPCServer:   gRPCServer,
		DBStorage:    sqliteStore,
		RedisStorage: reidsStore,
	}
}
