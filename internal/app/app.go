package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	grpcapp "github.com/Grino777/sso/internal/app/grpc"
	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models"
	"github.com/Grino777/sso/internal/services/auth"
	"github.com/Grino777/sso/internal/services/jwks"
	redisApp "github.com/Grino777/sso/internal/storage/redis"
	dbApp "github.com/Grino777/sso/internal/storage/sqlite"
	storageU "github.com/Grino777/sso/internal/utils/storage"
)

type Storage interface {
	Close() error
	GetApp(ctx context.Context, appID uint32) (app *models.App, err error)
	GetUser(ctx context.Context, username string) (*models.User, error)
	IsAdmin(ctx context.Context, user *models.User) (bool, error)
	SaveUser(ctx context.Context, user *models.User, passHash string) error
	SaveUserToken(ctx context.Context, user models.User, app models.App) error
}

type CacheStorage interface {
	Close() error
	GetApp(ctx context.Context, appID uint32) (*models.App, error)
	GetUser(ctx context.Context, user *models.User, app *models.App) (*models.User, error)
	IsAdmin(ctx context.Context, user *models.User, app *models.App) (bool, error)
	SaveApp(ctx context.Context, app *models.App) error
	SaveUser(ctx context.Context, user *models.User, app *models.App) error
}

type SSOApp struct {
	Config       *config.Config
	Logger       *slog.Logger
	GRPCServer   *grpcapp.GRPCApp
	DBStorage    Storage
	RedisStorage CacheStorage
}

type AppServices struct {
	jwksService *jwks.JwksService
	authService *auth.AuthService
}

type AuthConfig struct {
	Log             *slog.Logger
	DB              Storage
	Cache           CacheStorage
	TokenTTL        time.Duration
	RefreshTokenTTL time.Duration
	JwksService     *jwks.JwksService
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

	app := &SSOApp{Logger: log}

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

func (a *SSOApp) Stop() {
	const op = "app.Stop"

	log := a.Logger.With("op", op)

	a.DBStorage.Close()
	log.Debug("db session closed")

	a.RedisStorage.Close()
	log.Debug("redis session closed")
}

func initDB(a *SSOApp) error {
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

func initCache(a *SSOApp) error {
	const op = "app.initCache"

	redis, err := redisApp.New(a.Config.Redis, a.Logger)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	a.RedisStorage = redis
	a.Logger.Debug("cache initialized successfully", slog.String("op", op))
	return nil
}

func initGRPCServer(a *SSOApp, s *AppServices) {
	grpcServer := grpcapp.New(a.Logger, s, a.DBStorage, a.RedisStorage, a.Config)
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
		Log:             a.Logger,
		DB:              a.DBStorage,
		Cache:           a.RedisStorage,
		TokenTTL:        a.Config.TokenTTL,
		RefreshTokenTTL: a.Config.RefreshTokenTTL,
		JwksService:     jwksService,
	}

	authService := auth.New(authConfigs)

	return &AppServices{
		jwksService: jwksService,
		authService: authService,
	}
}

func initJwksService(a *SSOApp) (*jwks.JwksService, error) {
	const op = "app.initJwksService"

	jwksService, err := jwks.New(a.Logger, a.Config.KeysDir, a.Config.TokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}
	return jwksService, nil
}

func loadConfig(a *SSOApp) error {
	const op = "app.loadConfig"

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	a.Config = cfg
	a.Logger.Debug("config loaded successfully", slog.String("op", op))
	return nil
}
