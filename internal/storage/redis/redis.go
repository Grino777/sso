package redis

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	mu         sync.RWMutex
	cfg        config.RedisConfig
	client     *redis.Client
	maxRetries int
	retryDelay time.Duration // Задержка перед переподключением
	log        *slog.Logger
}

func New(cfg config.RedisConfig, log *slog.Logger) *RedisStorage {
	store := &RedisStorage{
		cfg:        cfg,
		maxRetries: 5,
		retryDelay: 4 * time.Second,
		log:        log,
	}

	if err := store.connectWithRetry(store.maxRetries); err != nil {
		panic("failed to connect to redis")
	}

	go store.listenConnection()

	return store
}

func (rs *RedisStorage) connect() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.client != nil {
		rs.client.Close()
	}

	options := &redis.Options{
		Addr:         rs.cfg.Addr,
		Password:     rs.cfg.Password,
		Username:     rs.cfg.User,
		DB:           rs.cfg.DB,
		MaxRetries:   rs.cfg.MaxRetries,
		DialTimeout:  rs.cfg.DialTimeout,
		ReadTimeout:  rs.cfg.Timeout,
		WriteTimeout: rs.cfg.Timeout,
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), rs.retryDelay)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return err
	}

	rs.client = client

	return nil

}

func (rs *RedisStorage) connectWithRetry(maxAttempts int) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := rs.connect(); err == nil {
			rs.log.Info("redis connection established")
			return nil
		} else {
			lastErr = err
			rs.log.Warn("failed to connect to redis",
				"atempt", attempt,
				"error", lastErr,
				"retryAfter", rs.retryDelay,
			)
			if attempt < maxAttempts {
				time.Sleep(rs.retryDelay)
			}
		}
	}
	return fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, lastErr)
}

func (rs *RedisStorage) listenConnection() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rs.mu.Lock()
		client := rs.client
		rs.mu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), rs.retryDelay)
		if client != nil {
			if err := client.Ping(ctx).Err(); err != nil {
				rs.connectWithRetry(rs.maxRetries)
			}
		} else {
			panic("redis client is nil")

		}

		cancel()
	}
}

func (rs *RedisStorage) withClient(ctx context.Context, fn func(*redis.Client) error) error {
	rs.mu.RLock()
	client := rs.client
	rs.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	if err := client.Ping(ctx).Err(); err != nil {
		rs.log.Warn("redis connection lost, attempting reconnect", "error", err)
		if err := rs.connectWithRetry(3); err != nil {
			return fmt.Errorf("failed to reconnect: %w", err)
		}
		rs.mu.RLock()
		client = rs.client
		rs.mu.RUnlock()
	}

	return fn(client)
}

func (rs *RedisStorage) SaveUser(ctx context.Context, username string, passHash string) error {
	return rs.withClient(ctx, func(client *redis.Client) error {
		return client.Set(ctx, username, passHash, 0).Err()
	})
}

func (rs *RedisStorage) GetUser(ctx context.Context, username string, appID uint32) (models.User, error) {
	panic("implemet me!")
}

func (rs *RedisStorage) GetApp(ctx context.Context, appID uint32) (app models.App, err error) {
	panic("implemet me!")
}

func (rs *RedisStorage) IsAdmin(ctx context.Context, username string) (isAdmin bool, err error) {
	panic("implemet me!")
}

// FIXME
func (rs *RedisStorage) SaveUserToken(ctx context.Context,
	user models.User,
	app models.App,
) error {
	panic("implemet me!")
}

// Close Redis session
func (rs *RedisStorage) Close() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.client != nil {
		rs.client.Close()
	}

	return nil
}
