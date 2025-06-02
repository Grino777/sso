package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/redis/go-redis/v9"
)

func (rs *RedisStorage) Connect(ctx context.Context) error {
	const op = opRedis + "Connect"

	log := rs.Logger.With(slog.String("op", op))

	if err := rs.connectWithRetry(rs.Cfg.MaxRetries); err != nil {
		log.Error("failed to connect to Redis after retries", logger.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Debug("redis connection established")
	return nil
}

func (rs *RedisStorage) connectWithRetry(maxAttempts int) error {
	const op = opRedis + "connectWithRetry"

	log := rs.Logger.With(slog.String("op", op))

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		rs.Mu.Lock()
		if rs.Client != nil {
			if err := rs.Client.Close(); err != nil {
				log.Warn("failed to close old Redis client", logger.Error(err))
			}
			rs.Client = nil
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
		rs.Mu.Unlock()

		// Проверка соединения
		ctx, cancel := context.WithTimeout(context.Background(), rs.RetryDelay)
		if err := client.Ping(ctx).Err(); err != nil {
			cancel() // Закрываем контекст сразу после ошибки
			rs.Mu.Lock()
			if closeErr := client.Close(); closeErr != nil {
				log.Warn("failed to close Redis client after failed ping", logger.Error(closeErr))
			}
			rs.Mu.Unlock()
			lastErr = err
			log.Warn("failed to connect to Redis",
				slog.Int("attempt", attempt),
				slog.Any("error", lastErr),
				slog.Duration("retryAfter", rs.RetryDelay),
			)
			if attempt < maxAttempts {
				time.Sleep(rs.RetryDelay)
			}
			continue
		}
		cancel() // Закрываем контекст после успешного пинга

		rs.Mu.Lock()
		rs.Client = client
		rs.Mu.Unlock()

		return nil
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, lastErr)
}

func (rs *RedisStorage) Close(ctx context.Context) error {
	rs.Mu.Lock()
	defer rs.Mu.Unlock()

	if rs.Client != nil {
		if err := rs.Client.Close(); err != nil {
			if err.Error() != "redis: client is closed" {
				return fmt.Errorf("failed to close redis client: %w", err)
			}
		}
		rs.Client = nil
	}
	rs.Logger.Debug("redis connection is closed")
	return nil
}

func withClient[T any](
	ctx context.Context,
	rs *RedisStorage,
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
		rs.Logger.Warn("redis connection lost, attempting reconnect", "error", err)
		if err := rs.connectWithRetry(rs.Cfg.MaxRetries); err != nil {
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
