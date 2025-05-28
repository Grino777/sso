package app

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/Grino777/sso/internal/app/apiserver"
	grpcapp "github.com/Grino777/sso/internal/app/grpc"
	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models/interfaces"
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

const opApp = "app."

type SSOApp struct {
	Config       *config.Config
	Logger       *slog.Logger
	GRPCServer   *grpcapp.GRPCApp
	Storage      interfaces.Storage
	CacheStorage interfaces.CacheStorage
	ApiServer    *apiserver.APIServer
	cancel       context.CancelFunc
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

func NewApp(
	ctx context.Context,
	logger *slog.Logger,
	cancel context.CancelFunc,
) (*SSOApp, error) {
	const op = opApp + "New"

	app := &SSOApp{Logger: logger, cancel: cancel}

	loadConfig(app)

	if err := initDB(ctx, app); err != nil {
		return app, fmt.Errorf("%s: %v", op, err)
	}
	if err := initCache(ctx, app); err != nil {
		return app, fmt.Errorf("%s: %v", op, err)
	}

	services := initServices(app)
	initGRPCServer(app, services)
	initApiServer(app)

	return app, nil
}

func (a *SSOApp) Run(ctx context.Context) {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Add(1)
	go func() {
		if err := a.GRPCServer.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		if err := a.ApiServer.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		redisApp.MonitorRedisConnection(ctx, a.CacheStorage, a.Logger, errChan)
		wg.Done()
	}()

	go func() {
		select {
		case <-ctx.Done():
			a.Stop()
		case err := <-errChan:
			a.Logger.Error("server failed", logger.Error(err))
			a.Stop()
		}
	}()

	wg.Wait()
}

func (a *SSOApp) Stop() {
	const op = "app.Stop"

	log := a.Logger.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.Storage.Close(ctx); err != nil {
		log.Error("failed to close db session", logger.Error(err))
	}
	if err := a.CacheStorage.Close(ctx); err != nil {
		log.Error("failed to close redis session", logger.Error(err))
	}

	a.cancel()
	log.Debug("application stopped")
}

func initApiServer(a *SSOApp) {
	server := apiserver.NewApiServer(a.Logger, a.Config.ApiServer)
	a.ApiServer = server
}

func initDB(ctx context.Context, a *SSOApp) error {
	const op = "grpc.app.initDb"

	log := a.Logger.With(slog.String("op", op))

	var db interfaces.Storage
	var err error

	switch a.Config.Database.DBType {
	case DBTypePostgres:
		db, err = postgres.NewPostgresStorage(ctx, a.Config.Database, a.Logger)
		if err != nil {
			log.Error(
				"failed to initialize Postgres storage",
				logger.Error(err),
			)
			a.Stop()
		}
	case DBTypeSQLite:
		if err := storageU.CheckStorageFolder(); err != nil {
			a.Logger.Error(
				"failed to check storage folder",
				logger.Error(err),
			)
			a.Stop()
		}
		db, err = dbApp.New("sqlite3", a.Config.Database.LocalStoragePath, a.Config.SuperUser, a.Logger)
		if err != nil {
			a.Logger.Error(
				"failed to initialize SQLite storage",
				logger.Error(err),
			)
			a.Stop()
		}
	default:
		a.Logger.Error(
			"unknown database type",
			logger.Error(err),
			slog.String("db_type", a.Config.Database.DBType),
		)
		a.Stop()
	}
	a.Storage = db
	log.Debug("db initialized successfully")
	return nil
}

func initCache(ctx context.Context, a *SSOApp) error {
	const op = "app.initCache"

	log := a.Logger.With(slog.String("op", op))

	// FIXME
	redis, err := redisApp.NewRedisStorage(ctx, a.Config.Redis, a.Logger)
	if err != nil {
		log.Error("cache initialized failed: %v", logger.Error(err))
	}
	a.CacheStorage = redis
	log.Debug("cache initialized successfully")
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
			logger.Error(err),
		)
		os.Exit(1)
	}

	authConfigs := auth.AuthService{
		Logger:      a.Logger,
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

func loadConfig(a *SSOApp) {
	const op = "app.loadConfig"

	log := a.Logger.With(slog.String("op", op))

	cfg, err := config.Load()
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			// // Вывод справки по флагам
			// fmt.Fprintf(os.Stderr, "Использование %s:\n", os.Args[0])
			config.GetFlagSet().Usage()
			os.Exit(0)
		}
		if errors.Is(err, config.ErrModeFlag) {
			log.Error(
				"invalid mode flag",
				logger.Error(err),
				slog.String("mode", config.GetModeFlag()),
			)
			config.GetFlagSet().Usage()
			os.Exit(2) // Код 2 для ошибок флагов
		} else if errors.Is(err, config.ErrDbFlag) {
			log.Error(
				"invalid db flag",
				logger.Error(err),
				slog.String("db", config.GetDbFlag()),
			)
			config.GetFlagSet().Usage()
			os.Exit(2)
		} else if os.IsNotExist(err) {
			log.Error(
				"config file not found",
				logger.Error(err),
			)
			os.Exit(3) // Код 3 для ошибок файлов
		} else {
			log.Error(
				"configuration loading error",
				logger.Error(err),
			)
			os.Exit(1) // Общий код для прочих ошибок
		}
		// Логирование остальных ошибок
		log.Error(
			"configuration loading error",
			logger.Error(err),
		)
		os.Exit(1)
	}
	a.Config = cfg
	log.Debug("configuration successfully loaded")
}
