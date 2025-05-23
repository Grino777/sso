package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models"

	"github.com/redis/go-redis/v9"
)

var (
	ErrCacheNotFound = errors.New("data not cached")
)

const opRedis = "storage.redis.redis."

type RedisStorage struct {
	Mu         sync.RWMutex
	Cfg        config.RedisConfig
	Client     *redis.Client
	MaxRetries int
	RetryDelay time.Duration // Задержка перед переподключением
	Log        *slog.Logger
}

func NewCacheStorage(cfg config.RedisConfig, log *slog.Logger) (*RedisStorage, error) {
	const op = opRedis + "NewCacheStorage"

	store := &RedisStorage{
		Cfg:        cfg,
		MaxRetries: 5,
		RetryDelay: 4 * time.Second,
		Log:        log,
	}

	if err := store.connectWithRetry(store.MaxRetries); err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	go store.listenConnection()

	return store, nil
}

// -----------------------------------User Block-----------------------------------

func (rs *RedisStorage) SaveUser(
	ctx context.Context,
	user models.User,
	appID uint32,
) (models.User, error) {
	const op = opRedis + "SaveUser"

	return withClient(ctx, rs, func(rc *redis.Client) (models.User, error) {
		user.Password = ""

		data, err := json.Marshal(user)
		if err != nil {
			return models.User{}, err
		}

		resString := fmt.Sprintf("users:%d:%s", appID, user.Username)
		err = rc.Set(ctx, resString, data, rs.Cfg.TokenTTL).Err()
		if err != nil {
			return models.User{}, fmt.Errorf("%s: %v", op, err)
		}
		rs.Log.Debug("user successfuly cached", "username", user.Username)
		return models.User{}, nil
	})
}

func (rs *RedisStorage) GetUser(
	ctx context.Context,
	username string,
	appID uint32,
) (models.User, error) {
	const op = opRedis + "GetUser"

	key := fmt.Sprintf("users:%d:%s", appID, username)
	result, err := withClient(ctx, rs, func(rc *redis.Client) (string, error) {
		return rc.Get(ctx, key).Result()
	})
	if err != nil {
		if err == redis.Nil {
			return models.User{}, fmt.Errorf("%s: %w for username %s and appID %d", op, ErrCacheNotFound, username, appID)
		}
		return models.User{}, fmt.Errorf("%s: failed to get user: %w", op, err)
	}

	var user models.User
	if err := json.Unmarshal([]byte(result), &user); err != nil {
		return models.User{}, fmt.Errorf("%s: failed to unmarshal user: %w", op, err)
	}

	return user, nil
}

// FIXME
func (rs *RedisStorage) IsAdmin(
	ctx context.Context,
	user models.User,
	app models.App,
) (bool, error) {
	panic("implement me!")
}

// -----------------------------------End Block------------------------------------

// -----------------------------------App Block------------------------------------

func (rs *RedisStorage) GetApp(
	ctx context.Context,
	appID uint32,
) (models.App, error) {
	const op = opRedis + "GetApp"

	key := fmt.Sprintf("apps:%d", appID)
	result, err := withClient(ctx, rs, func(rc *redis.Client) (string, error) {
		return rc.Get(ctx, key).Result()
	})
	if err != nil {
		if err == redis.Nil {
			return models.App{}, fmt.Errorf("%s: %w for appID %d", op, ErrCacheNotFound, appID)
		}
		return models.App{}, fmt.Errorf("%s: failed to get app: %w", op, err)
	}

	var app models.App
	if err := json.Unmarshal([]byte(result), &app); err != nil {
		return models.App{}, fmt.Errorf("%s: failed to unmarshal user: %w", op, err)
	}

	return app, nil
}

func (rs *RedisStorage) SaveApp(
	ctx context.Context,
	app models.App,
) error {
	const op = opRedis + "SaveApp"

	key := fmt.Sprintf("apps:%d", app.ID)
	_, err := withClient(ctx, rs, func(rc *redis.Client) (models.App, error) {
		data, err := json.Marshal(app)
		if err != nil {
			return models.App{}, fmt.Errorf("%s: %v", op, err)
		}

		_, err = rc.Set(ctx, key, data, 0).Result()
		if err != nil {
			return models.App{}, fmt.Errorf("%s: %v", op, err)
		}

		rs.Log.Info("app added to cache", "appID", app.ID)
		return models.App{}, nil
	})
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

// -----------------------------------End Block------------------------------------

// Close Redis session
func (rs *RedisStorage) Close(ctx context.Context) error {
	rs.Mu.Lock()
	defer rs.Mu.Unlock()

	if rs.Client != nil {
		rs.Client.Close()
	}
	return nil
}
