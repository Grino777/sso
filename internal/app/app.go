package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	grpcapp "github.com/Grino777/sso/internal/app/grpc"
	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models/interfaces"
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

type SSOApp struct {
	Ctx          context.Context
	Config       *config.Config
	Logger       *slog.Logger
	GRPCServer   *grpcapp.GRPCApp
	Storage      interfaces.Storage
	CacheStorage interfaces.CacheStorage
}

type AppServices struct {
	jwksService *jwks.JwksService
	authService *auth.AuthService
}

type AuthConfig struct {
	Log         *slog.Logger
	DB          interfaces.Storage
	Cache       interfaces.CacheStorage
	Tokens      config.TokenConfig
	JwksService *jwks.JwksService
}

func (s *AppServices) Auth() *auth.AuthService {
	return s.authService
}

func (s *AppServices) Jwks() *jwks.JwksService {
	return s.jwksService
}

func New(
	log *slog.Logger,
) (*SSOApp, error) {
	const op = "app.New"

	app := &SSOApp{Ctx: context.Background(), Logger: log}

	if err := loadConfig(app); err != nil {
		if err == flag.ErrHelp {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	if err := initDB(app); err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	if err := initCache(app); err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	services := initServices(app)
	initGRPCServer(app, services)

	return app, nil
}

func (a *SSOApp) Stop() {
	const op = "app.Stop"

	log := a.Logger.With(slog.String("op", op))

	if err := a.Storage.Close(a.Ctx); err != nil {
		log.Error("failed to close db session", slog.String("error", err.Error()))
	} else {
		log.Debug("db session closed")
	}

	if err := a.CacheStorage.Close(a.Ctx); err != nil {
		log.Error("failed to close redis session", slog.String("error", err.Error()))
	} else {
		log.Debug("redis session closed")
	}
}

func initDB(a *SSOApp) error {
	const op = "grpc.app.initDb"

	var db interfaces.Storage
	var err error

	switch a.Config.Database.DBType {
	case DBTypePostgres:
		db, err = postgres.NewPostgresStorage(a.Ctx, a.Config.Database)
		if err != nil {
			a.Logger.Error(
				"failed to initialize Postgres storage",
				slog.String("op", op),
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
	case DBTypeSQLite:
		if err := storageU.CheckStorageFolder(); err != nil {
			a.Logger.Error(
				"failed to check storage folder",
				slog.String("op", op),
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
		db, err = dbApp.New("sqlite3", a.Config.Database.LocalStoragePath, a.Config.SuperUser)
		if err != nil {
			a.Logger.Error(
				"failed to initialize SQLite storage",
				slog.String("op", op),
				slog.String("error", err.Error()),
			)
			os.Exit(1)
		}
	default:
		a.Logger.Error(
			"unknown database type",
			slog.String("op", op),
			slog.String("db_type", a.Config.Database.DBType),
		)
		os.Exit(1)
	}

	a.Storage = db
	a.Logger.Debug("db initialized successfully", slog.String("op", op))
	return nil
}

func initCache(a *SSOApp) error {
	const op = "app.initCache"

	redis, err := redisApp.NewCacheStorage(a.Config.Redis, a.Logger)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	a.CacheStorage = redis
	a.Logger.Debug("cache initialized successfully", slog.String("op", op))
	return nil
}

func initGRPCServer(a *SSOApp, s *AppServices) {
	grpcServer := grpcapp.New(a.Logger, s, a.Storage, a.CacheStorage, a.Config)
	a.GRPCServer = grpcServer
	a.Logger.Debug("gRPC server initialized")
}

func initServices(a *SSOApp) *AppServices {
	const op = "app.initServices"

	jwksService, err := initJwksService(a)
	if err != nil {
		a.Logger.Warn(
			"jwks service not initialized",
			slog.String("op", op),
			slog.String("error", err.Error()),
		)
		panic("jwks service not initialized")
	}

	authConfigs := auth.AuthService{
		Log:         a.Logger,
		DB:          a.Storage,
		Cache:       a.CacheStorage,
		Tokens:      a.Config.Tokens,
		JwksService: jwksService,
	}

	authService := auth.New(authConfigs)

	return &AppServices{
		jwksService: jwksService,
		authService: authService,
	}
}

func initJwksService(a *SSOApp) (*jwks.JwksService, error) {
	const op = "app.initJwksService"

	jwksService, err := jwks.New(a.Logger, a.Config.FS.KeysDir, a.Config.Tokens.TokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	return jwksService, nil
}

func loadConfig(a *SSOApp) error {
	const op = "app.loadConfig"

	cfg, err := config.Load()
	if err != nil {
		if err == flag.ErrHelp {
			// Вывод справки по флагам
			fmt.Fprintf(os.Stderr, "Использование %s:\n", os.Args[0])
			config.GetFlagSet().Usage()
			os.Exit(0)
		}
		if errors.Is(err, config.ErrModeFlag) {
			a.Logger.Error(
				"invalid mode flag",
				slog.String("op", op),
				slog.String("error", err.Error()),
				slog.String("mode", config.GetModeFlag()),
			)
			config.GetFlagSet().Usage()
			os.Exit(2) // Код 2 для ошибок флагов
		} else if errors.Is(err, config.ErrDbFlag) {
			a.Logger.Error(
				"invalid db flag",
				slog.String("op", op),
				slog.String("error", err.Error()),
				slog.String("db", config.GetDbFlag()),
			)
			config.GetFlagSet().Usage()
			os.Exit(2)
		} else if os.IsNotExist(err) {
			a.Logger.Error(
				"config file not found",
				slog.String("op", op),
				slog.String("error", err.Error()),
			)
			os.Exit(3) // Код 3 для ошибок файлов
		} else {
			a.Logger.Error(
				"configuration loading error",
				slog.String("op", op),
				slog.String("error", err.Error()),
			)
			os.Exit(1) // Общий код для прочих ошибок
		}
		// Логирование остальных ошибок
		a.Logger.Error(
			"configuration loading error",
			slog.String("op", op),
			slog.String("error", err.Error()),
		)
		os.Exit(1)
	}

	a.Config = cfg
	a.Logger.Debug("configuration successfully loaded", slog.String("op", op))
	return nil
}
