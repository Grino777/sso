package storage

import (
	"context"

	"github.com/Grino777/sso/internal/domain/models"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage/storage_mock.go -package=mocks_storage
type Storage interface {
	StorageUserProvider
	StorageAppProvider
	StorageTokenProvider
	Connector
}

type StorageUserProvider interface {
	SaveUser(ctx context.Context, user, passHash string) error
	GetUser(ctx context.Context, username string) (models.User, error)
	IsAdmin(ctx context.Context, username string) (bool, error)
}

type StorageAppProvider interface {
	GetApp(ctx context.Context, appID uint32) (models.App, error)
}

type StorageTokenProvider interface {
	DeleteRefreshToken(ctx context.Context, userID uint64, appID uint32, token models.Token) error
	SaveRefreshToken(ctx context.Context, userID uint64, appID uint32, token models.Token) error
}

type Connector interface {
	Connect(ctx context.Context) error
	Close(ctx context.Context) error
}
