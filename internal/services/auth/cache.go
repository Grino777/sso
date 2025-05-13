package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Grino777/sso/internal/domain/models"
	"github.com/Grino777/sso/internal/storage"
	"github.com/Grino777/sso/internal/storage/redis"
)

func GetCachedApp(ctx context.Context,
	db DBStorage,
	cache CacheStorage,
	appID uint32,
) (app *models.App, err error) {
	const op = "services.auth.GetCachedApp"

	app, err = cache.GetApp(ctx, appID)
	if err != nil {
		if !errors.Is(err, redis.ErrCacheNotFound) {
			return nil, err
		}
	}

	if app == nil {
		app, err = db.GetApp(ctx, appID)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
			}
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		cache.SaveApp(ctx, app)
	}
	return app, err
}

func GetCachedUser(
	ctx context.Context,
	db DBStorage,
	cache CacheStorage,
	userObj *models.User,
	app *models.App,
) (user *models.User, err error) {
	const op = "services.auth.GetCachedUser"

	user, err = cache.GetUser(ctx, userObj, app)
	if err != nil {
		if !errors.Is(err, redis.ErrCacheNotFound) {
			return nil, err
		}
	}

	if user == nil {
		user, err = db.GetUser(ctx, userObj.Username)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				return nil, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
			}
			return nil, fmt.Errorf("%s: %w", op, err)
		}
	}
	return user, err
}
