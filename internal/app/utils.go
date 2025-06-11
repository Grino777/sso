package app

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/Grino777/sso/internal/app/admin"
	grpcapp "github.com/Grino777/sso/internal/app/grpc"
	"github.com/Grino777/sso/internal/config"
	storageI "github.com/Grino777/sso/internal/interfaces/storage"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/Grino777/sso/internal/services/auth"
	"github.com/Grino777/sso/internal/services/jwks"
	"github.com/Grino777/sso/internal/storage/postgres"
	redisApp "github.com/Grino777/sso/internal/storage/redis"
	dbApp "github.com/Grino777/sso/internal/storage/sqlite"
	storageU "github.com/Grino777/sso/internal/utils/storage"
)

const (
	DBTypePostgres = "postgres"
	DBTypeSQLite   = "sqlite"
)

func (a *SSOApp) initApiServer() {
	server := admin.NewApiServer(a.Logger, a.Config.ApiServer)
	a.Apps.Api = server
	a.Logger.Debug("api server successfully initialized")
}

func (a *SSOApp) initDB() error {
	const op = "grpc.app.initDb"

	log := a.Logger.With(slog.String("op", op))

	var db storageI.Storage
	var err error

	switch a.Config.Database.DBType {
	case DBTypePostgres:
		db = postgres.NewPostgresStorage(a.Config.Database, a.Logger)
	case DBTypeSQLite:
		if err := storageU.CheckStorageFolder(); err != nil {
			a.Logger.Error(
				"failed to check storage folder",
				logger.Error(err),
			)
		}
		db = dbApp.New("sqlite3", a.Config.Database.LocalStoragePath, a.Config.SuperUser, a.Logger)
	default:
		a.Logger.Error(
			"unknown database type",
			logger.Error(err),
			slog.String("db_type", a.Config.Database.DBType),
		)
	}
	a.Storages.Db = db
	log.Debug("db initialized successfully")
	return nil
}

func (a *SSOApp) initCache() error {
	const op = "app.initCache"

	log := a.Logger.With(slog.String("op", op))

	// FIXME
	redis := redisApp.NewRedisStorage(a.Logger, a.Config.Redis, a.internal.errChan)

	a.Storages.Cache = redis
	log.Debug("cache initialized successfully", slog.String("addr", a.Config.Redis.Addr))
	return nil
}

func (a *SSOApp) initGRPCApp(s *GrpcServices) {
	grpcApp := grpcapp.NewGrpcApp(a.Logger, s, a.Config)
	a.Apps.Grpc = grpcApp
	a.Logger.Debug("gRPC server successfully initialized")
}

func (a *SSOApp) initServices() *GrpcServices {
	const op = "app.initServices"

	jwksService, err := a.initJwksService()
	if err != nil {
		a.Logger.Warn(
			"jwks service not initialized",
			slog.String("op", op),
			logger.Error(err),
		)
		os.Exit(1)
	}

	authConfigs := auth.AuthService{
		Logger:      a.Logger,
		DB:          a.Storages.Db,
		Cache:       a.Storages.Cache,
		Tokens:      a.Config.Tokens,
		JwksService: jwksService,
	}

	authService := auth.NewAuthService(authConfigs)
	a.Logger.Debug("all services successfully initialized")

	return &GrpcServices{
		jwksService: jwksService,
		authService: authService,
	}
}

func (a *SSOApp) initJwksService() (*jwks.JwksService, error) {
	const op = "app.initJwksService"

	jwksService, err := jwks.NewJwksService(a.Logger, a.Config.FS.KeysDir, a.Config.Tokens.TokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	a.Logger.Debug("jwks service successfully initialized")

	return jwksService, nil
}

func (a *SSOApp) loadConfig() error {
	const op = "app.loadConfig"

	log := a.Logger.With(slog.String("op", op))

	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			config.GetFlagSet().Usage()
			os.Exit(1)
		}
		if errors.Is(err, config.ErrModeFlag) {
			log.Error(
				"invalid mode flag",
				logger.Error(err),
				slog.String("mode", config.GetModeFlag()),
			)
			config.GetFlagSet().Usage()
			return fmt.Errorf("%s: %w", op, err)
		} else if errors.Is(err, config.ErrDbFlag) {
			log.Error(
				"invalid db flag",
				logger.Error(err),
				slog.String("db", config.GetDbFlag()),
			)
			config.GetFlagSet().Usage()
			return fmt.Errorf("%s: %w", op, err)
		} else if os.IsNotExist(err) {
			log.Error(
				"config file not found",
				logger.Error(err),
			)
			return fmt.Errorf("%s: %w", op, err)
		} else {
			log.Error(
				"configuration loading error",
				logger.Error(err),
			)
			return fmt.Errorf("%s: %w", op, err)
		}
	}
	a.Config = cfg
	log.Debug("configuration successfully loaded")
	return nil
}
