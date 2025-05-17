package app

import (
	"fmt"
	"log/slog"

	grpcapp "github.com/Grino777/sso/internal/app/grpc"
	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/services/auth"
	"github.com/Grino777/sso/internal/services/jwks"
	redisApp "github.com/Grino777/sso/internal/storage/redis"
	dbApp "github.com/Grino777/sso/internal/storage/sqlite"
	storageU "github.com/Grino777/sso/internal/utils/storage"
)

type App struct {
	Config       *config.Config
	Logger       *slog.Logger
	GRPCServer   *grpcapp.GRPCApp
	DBStorage    *dbApp.SQLiteStorage
	RedisStorage *redisApp.RedisStorage
}

type Services struct {
	jwksService *jwks.JwksService
	authService *auth.AuthService
}

func (s *Services) Auth() *auth.AuthService {
	return s.authService
}

func (s *Services) Jwks() *jwks.JwksService {
	return s.jwksService
}

func New(
	log *slog.Logger,
) (*App, error) {
	const op = "app.New"

	app := &App{Logger: log}

	if err := loadConfig(app); err != nil {
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

func (a *App) Stop() {
	const op = "app.Stop"

	log := a.Logger.With("op", op)

	a.DBStorage.Close()
	log.Debug("db session closed")

	a.RedisStorage.Close()
	log.Debug("redis session closed")
}

func initDB(a *App) error {
	const op = "grpc.app.initDb"

	if err := storageU.CheckStorageFolder(); err != nil {
		return err
	}

	db, err := dbApp.New("sqlite3", a.Config.DB.StoragePath, a.Config.DBUser)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	a.DBStorage = db
	a.Logger.Debug("db initialized successfully", slog.String("op", op))
	return nil
}

func initCache(a *App) error {
	const op = "app.initCache"

	redis, err := redisApp.New(a.Config.Redis, a.Logger)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	a.RedisStorage = redis
	a.Logger.Debug("cache initialized successfully", slog.String("op", op))
	return nil
}

func initGRPCServer(a *App, s *Services) {
	grpcServer := grpcapp.New(a.Logger, s, a.DBStorage, a.RedisStorage, a.Config)
	a.GRPCServer = grpcServer
	a.Logger.Debug("gRPC server initialized")
}

func initServices(a *App) *Services {
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

	authService := auth.New(a.Logger, a.DBStorage, a.RedisStorage, a.Config.TokenTTL, jwksService)

	return &Services{
		jwksService: jwksService,
		authService: authService,
	}
}

func initJwksService(a *App) (*jwks.JwksService, error) {
	const op = "app.initJwksService"

	jwksService, err := jwks.New(a.Logger, a.Config.KeysDir, a.Config.TokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	return jwksService, nil
}

func loadConfig(a *App) error {
	const op = "app.loadConfig"

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	a.Config = cfg
	a.Logger.Debug("config loaded successfully", slog.String("op", op))
	return nil
}
