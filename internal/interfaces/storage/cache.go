package storage

import (
	"context"

	"github.com/Grino777/sso/internal/domain/models"
)

//go:generate mockgen -source=cache.go -destination=mocks/cache/redis_storage_mock.go -package=mock_cache
type CacheStorage interface {
	CacheUserProvider
	CacheAppProvider
	CacheConnector
}

type CacheUserProvider interface {
	GetUser(ctx context.Context, username string, appID uint32) (models.User, error)
	SaveUser(ctx context.Context, user models.User, appID uint32) (models.User, error)
	IsAdmin(ctx context.Context, user models.User, app models.App) (bool, error)
}

type CacheAppProvider interface {
	GetApp(ctx context.Context, appID uint32) (models.App, error)
	SaveApp(ctx context.Context, app models.App) error
}

type CacheConnector interface {
	Connect(ctx context.Context, errChan chan<- error) error
	Close(ctx context.Context) error
}
