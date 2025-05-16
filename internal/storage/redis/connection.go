package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func (rs *RedisStorage) connect() error {
	rs.Mu.Lock()
	defer rs.Mu.Unlock()

	if rs.Client != nil {
		rs.Client.Close()
	}

	options := &redis.Options{
		Addr:         rs.Cfg.Addr,
		Password:     rs.Cfg.Password,
		Username:     rs.Cfg.User,
		DB:           rs.Cfg.DB,
		MaxRetries:   rs.Cfg.MaxRetries,
		DialTimeout:  rs.Cfg.DialTimeout,
		ReadTimeout:  rs.Cfg.Timeout,
		WriteTimeout: rs.Cfg.Timeout,
	}

	client := redis.NewClient(options)

	ctx, cancel := context.WithTimeout(context.Background(), rs.RetryDelay)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		return err
	}

	rs.Client = client

	return nil

}

func (rs *RedisStorage) connectWithRetry(maxAttempts int) error {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := rs.connect(); err == nil {
			rs.Log.Debug("redis connection established")
			return nil
		} else {
			lastErr = err
			rs.Log.Warn("failed to connect to redis",
				"atempt", attempt,
				"error", lastErr,
				"retryAfter", rs.RetryDelay,
			)
			if attempt < maxAttempts {
				time.Sleep(rs.RetryDelay)
			}
		}
	}
	return fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, lastErr)
}

func (rs *RedisStorage) listenConnection() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		rs.Mu.Lock()
		client := rs.Client
		rs.Mu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), rs.RetryDelay)
		if client != nil {
			if err := client.Ping(ctx).Err(); err != nil {
				rs.connectWithRetry(rs.MaxRetries)
			}
		} else {
			panic("redis client is nil")

		}

		cancel()
	}
}

func withClient[T any](
	rs *RedisStorage,
	ctx context.Context,
	fn func(*redis.Client) (T, error),
) (T, error) {
	var zero T

	rs.Mu.RLock()
	client := rs.Client
	rs.Mu.RUnlock()

	if client == nil {
		return zero, fmt.Errorf("redis client is not initialized")
	}

	if err := client.Ping(ctx).Err(); err != nil {
		rs.Log.Warn("redis connection lost, attempting reconnect", "error", err)
		if err := rs.connectWithRetry(3); err != nil {
			return zero, fmt.Errorf("failed to reconnect: %w", err)
		}
		rs.Mu.RLock()
		client = rs.Client
		rs.Mu.RUnlock()
	}

	result, err := fn(client)
	if err != nil {
		return zero, err
	}

	return result, nil
}
