package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Grino777/sso/internal/domain/models"
	"github.com/Grino777/sso/internal/storage"
	"github.com/Grino777/sso/internal/storage/redis"
)

const cacheOp = "services.auth.cache."

func (s *AuthService) GetCachedApp(
	ctx context.Context,
	appID uint32,
) (models.App, error) {
	const op = cacheOp + "GetCachedApp"

	app, err := s.Cache.GetApp(ctx, appID)
	if err != nil {
		if !errors.Is(err, redis.ErrCacheNotFound) {
			return app, fmt.Errorf("%s: %w", op, err)
		}
		app, err = s.DB.GetApp(ctx, appID)
		if err != nil {
			if errors.Is(err, storage.ErrAppNotFound) {
				return models.App{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
			}
			return models.App{}, fmt.Errorf("%s: %w", op, err)
		}
		s.Cache.SaveApp(ctx, app)
	}
	return app, nil
}

func (s *AuthService) GetCachedUser(
	ctx context.Context,
	username string,
	appID uint32,
) (user models.User, err error) {
	const op = cacheOp + "GetCachedUser"

	user, err = s.Cache.GetUser(ctx, username, appID)
	if err != nil {
		if !errors.Is(err, redis.ErrCacheNotFound) {
			return user, err
		}
		user, err = s.DB.GetUser(ctx, username)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				return models.User{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
			}
			return models.User{}, fmt.Errorf("%s: %w", op, err)
		}
		s.Cache.SaveUser(ctx, user, appID)
	}
	return user, err
}
