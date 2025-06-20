package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Grino777/sso/internal/config"
	storageI "github.com/Grino777/sso/internal/interfaces/storage"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/Grino777/sso/internal/services/auth"
	"github.com/Grino777/sso/internal/services/jwks"
	"github.com/Grino777/sso/internal/services/keys/store"
)

const opApp = "app."

//go:generate mockgen -source=app.go -destination=mocks/admin/admin_api_mock.go -package=admin
type adminServer interface {
	Run(ctx context.Context) error
	Stop() error
}

//go:generate mockgen -source=app.go -destination=mocks/grpc/grpc_app_mock.go -package=grpc
type grpcApp interface {
	Run(ctx context.Context) error
	Stop()
}

type Apps struct {
	Grpc grpcApp
	Api  adminServer
}

// Contains storages for App
type Storages struct {
	Db    storageI.Storage
	Cache storageI.CacheStorage
}

// Internal variables for App
type Internal struct {
	errChan chan error
	cancel  context.CancelFunc
}

type SSOApp struct {
	Config   *config.Config
	Logger   *slog.Logger
	Storages Storages
	Apps     Apps
	internal Internal
}

type GrpcServices struct {
	jwksService *jwks.JwksService
	authService *auth.AuthService
}

func (s *GrpcServices) Auth() *auth.AuthService {
	return s.authService
}

func (s *GrpcServices) Jwks() *jwks.JwksService {
	return s.jwksService
}

func NewApp(
	log *slog.Logger,
) (*SSOApp, error) {
	app := &SSOApp{Logger: log}

	if err := app.loadConfig(); err != nil {
		return app, err
	}

	// keysManager, err := km.NewKeysManager(log, app.Config.FS.KeysDir)
	// if err != nil {
	// 	log.Error("failed to create keys manager", logger.Error(err))
	// 	return nil, err
	// }

	keysStore, err := store.NewKeysStore(log, app.Config.Path, app.Config.TTL)
	if err != nil {
		log.Error("failed to create keys store", logger.Error(err))
		return nil, err
	}

	app.initDB()
	app.initCache()
	services := app.initServices(keysStore)
	app.initGRPCApp(services, keysStore)
	app.initApiServer(keysStore)

	return app, nil
}

func (a *SSOApp) Run(mainErrChan chan error) {
	const op = opApp + "Run"

	var wg sync.WaitGroup
	errChan := make(chan error, 4)

	log := a.Logger.With(slog.String("op", op))

	ctx, cancel := context.WithCancel(context.Background())
	a.internal.cancel = cancel

	if err := a.Storages.Db.Connect(ctx); err != nil {
		log.Error("failed to connect database")
		errChan <- err
		return
	}
	if err := a.Storages.Cache.Connect(ctx, errChan); err != nil {
		log.Error("failed to connect redis")
		errChan <- err
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.Apps.Grpc.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.Apps.Api.Run(ctx); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Debug("stop signal received, initiating shutdown")
	case err := <-errChan:
		a.Logger.Error("server failed", logger.Error(err))
		mainErrChan <- err
	}

	wg.Wait()
}

func (a *SSOApp) Stop() {
	const op = "app.Stop"

	a.internal.cancel()

	log := a.Logger.With(slog.String("op", op))
	db := a.Storages.Db
	cache := a.Storages.Cache

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.Close(ctx); err != nil {
		log.Error("failed to close db session", logger.Error(err))
	}
	if err := cache.Close(ctx); err != nil {
		log.Error("failed to close redis session", logger.Error(err))
	}

	log.Debug("application stopped")
}
