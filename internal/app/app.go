package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	grpcapp "github.com/Grino777/sso/internal/app/grpc"
	"github.com/Grino777/sso/internal/config"
	storageI "github.com/Grino777/sso/internal/interfaces/storage"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/Grino777/sso/internal/services/auth"
	"github.com/Grino777/sso/internal/services/jwks"
)

const opApp = "app."

type AdminApi interface {
	RegisterRoutes()
	Run(ctx context.Context) error
	Stop() error
}

type AdminService interface {
	RotateJwksService() error
}

type SSOApp struct {
	Config       *config.Config
	Logger       *slog.Logger
	GRPCServer   *grpcapp.GRPCApp
	Storage      storageI.Storage
	CacheStorage storageI.CacheStorage
	ApiServer    AdminApi
	cancel       context.CancelFunc
}

type AppServices struct {
	jwksService *jwks.JwksService
	authService *auth.AuthService
}

type AuthConfig struct {
	Log         *slog.Logger
	DB          storageI.Storage
	Cache       storageI.CacheStorage
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
	logger *slog.Logger,
) (*SSOApp, error) {
	app := &SSOApp{Logger: logger}

	if err := app.loadConfig(); err != nil {
		return app, err
	}

	app.initDB()
	app.initCache()

	return app, nil
}

func (a *SSOApp) Run(mainChan chan error) {
	const op = opApp + "Run"

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	log := a.Logger.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	a.cancel = cancel

	if err := a.Storage.Connect(ctx); err != nil {
		log.Error("failed to connect database")
		mainChan <- err
		return
	}
	if err := a.CacheStorage.Connect(ctx); err != nil {
		log.Error("failed to connect redis")
		mainChan <- err
		return
	}

	services := a.initServices()
	a.initGRPCServer(services)
	a.initApiServer()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.GRPCServer.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.ApiServer.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			log.Debug("stop signal recived")
		case err := <-errChan:
			a.Logger.Error("server failed", logger.Error(err))
			mainChan <- err
			a.cancel()
		}
	}()

	wg.Wait()
}

func (a *SSOApp) Stop() {
	const op = "app.Stop"

	log := a.Logger.With(slog.String("op", op))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if a.Storage != nil {
		if err := a.Storage.Close(ctx); err != nil {
			log.Error("failed to close db session", logger.Error(err))
		}
	}
	if a.CacheStorage != nil {
		if err := a.CacheStorage.Close(ctx); err != nil {
			log.Error("failed to close redis session", logger.Error(err))
		}
	}
	if a.ApiServer != nil {
		if err := a.ApiServer.Stop(); err != nil {
			log.Error("failed to stopping api server", logger.Error(err))
		}
	}
	if a.GRPCServer != nil {
		a.GRPCServer.Stop()
	}

	log.Debug("application stopped")
}
