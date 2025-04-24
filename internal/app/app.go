package app

import (
	"log/slog"

	grpcapp "github.com/Grino777/sso/internal/app/grpc"
	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/services/auth"
	"github.com/Grino777/sso/internal/storage"
	redisApp "github.com/Grino777/sso/internal/storage/redis"
	dbApp "github.com/Grino777/sso/internal/storage/sqlite"
)

type App struct {
	Config       *config.Config
	Logger       *slog.Logger
	GRPCServer   *grpcapp.GRPCApp
	DBStorage    storage.Storage
	RedisStorage storage.Storage
}

func New(
	cfg *config.Config,
	log *slog.Logger,
) *App {
	db := dbApp.New("sqlite3", cfg.DB.Storage_path, cfg.DBUser)
	dbStorage := dbApp.ToStorage(db)

	redis := redisApp.New(cfg.Redis, log)
	cacheStorage := redisApp.ToStorage(redis)

	authService := auth.New(log, dbStorage, cacheStorage, cfg.TokenTTL)

	grpcServer := grpcapp.New(log, authService, dbStorage, cacheStorage, int(cfg.GRPC.Port), cfg.Mode)

	return &App{
		Config:       cfg,
		Logger:       log,
		GRPCServer:   grpcServer,
		DBStorage:    dbStorage,
		RedisStorage: cacheStorage,
	}
}

func (a *App) Stop() {
	const op = "app.app.Stop"

	log := a.Logger.With("op", op)

	a.DBStorage.Closer.Close()
	log.Debug("db session closed")
	a.RedisStorage.Closer.Close()
	log.Debug("redis session closed")
}
