package storage

import (
	"context"

	"github.com/Grino777/sso/internal/domain/models"
)

type Storage struct {
	Users  UserStorage
	Apps   AppStorage
	Closer StorageCloser
}

type UserStorage interface {
	SaveUser(ctx context.Context, username string, passHash string) error
	GetUser(ctx context.Context, username string, appID uint32) (models.User, error)
	IsAdmin(ctx context.Context, username string) (bool, error)
	SaveUserToken(ctx context.Context, user models.User, app models.App) error
}

type AppStorage interface {
	GetApp(ctx context.Context, appID uint32) (models.App, error)
}

type StorageCloser interface {
	Close() error
}
