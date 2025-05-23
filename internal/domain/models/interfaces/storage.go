package interfaces

import (
	"context"

	"github.com/Grino777/sso/internal/domain/models"
)

type Storage interface {
	StorageUserProvider
	StorageAppProvider
	StorageTokenProvider
	Closer
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

type Closer interface {
	Close(ctx context.Context) error
}
