package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/redis/go-redis/v9"
)

func (rs *RedisStorage) Connect(ctx context.Context, errChan chan<- error) error {
	const op = opRedis + "Connect"

	log := rs.logger.With(slog.String("op", op))

	if err := rs.connectWithRetry(rs.cfg.MaxRetries); err != nil {
		log.Error("failed to connect to Redis after retries", logger.Error(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	rs.errChan = errChan
	log.Debug("redis connection established")

	return nil
}

func (rs *RedisStorage) connectWithRetry(maxAttempts int) error {
	const op = opRedis + "connectWithRetry"

	log := rs.logger.With(slog.String("op", op))

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		rs.mu.Lock()
		if rs.client != nil {
			if err := rs.client.Close(); err != nil {
				log.Warn("failed to close old Redis client", logger.Error(err))
			}
			rs.client = nil
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
		rs.mu.Unlock()

		// Проверка соединения
		ctx, cancel := context.WithTimeout(context.Background(), rs.cfg.DialTimeout)
		if err := client.Ping(ctx).Err(); err != nil {
			cancel() // Закрываем контекст сразу после ошибки
			rs.mu.Lock()
			if closeErr := client.Close(); closeErr != nil {
				log.Warn("failed to close Redis client after failed ping", logger.Error(closeErr))
			}
			rs.mu.Unlock()
			lastErr = err
			log.Warn("failed to connect to Redis",
				slog.Int("attempt", attempt),
				slog.Any("error", lastErr),
				slog.Duration("retryAfter", rs.cfg.RetryDelay),
			)
			if attempt < maxAttempts {
				time.Sleep(rs.cfg.RetryDelay)
			}
			continue
		}
		cancel() // Закрываем контекст после успешного пинга

		rs.mu.Lock()
		rs.client = client
		rs.mu.Unlock()

		return nil
	}

	return fmt.Errorf("failed to connect after %d attempts: %w", maxAttempts, lastErr)
}

func (rs *RedisStorage) Close(ctx context.Context) error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.client != nil {
		if err := rs.client.Close(); err != nil {
			if err.Error() != "redis: client is closed" {
				return fmt.Errorf("failed to close redis client: %w", err)
			}
		}
		rs.client = nil
	}
	rs.logger.Debug("redis connection is closed")
	return nil
}

func withClient[T any](
	ctx context.Context,
	rs *RedisStorage,
	fn func(*redis.Client) (T, error),
) (T, error) {
	var zero T

	rs.mu.RLock()
	client := rs.client
	errChan := rs.errChan
	rs.mu.RUnlock()

	if client == nil {
		if err := rs.connectWithRetry(rs.cfg.MaxRetries); err != nil {
			errChan <- fmt.Errorf("failed to reconnect: %w", err)
			return zero, fmt.Errorf("failed to reconnect: %w", err)
		}
		rs.mu.RLock()
		client = rs.client
		rs.mu.RUnlock()
	}
	if err := client.Ping(ctx).Err(); err != nil {
		rs.logger.Warn("redis connection lost, attempting reconnect", "error", err)
		if err := rs.connectWithRetry(rs.cfg.MaxRetries); err != nil {
			errChan <- err
			return zero, fmt.Errorf("failed to reconnect: %w", err)
		}
		rs.mu.RLock()
		client = rs.client
		rs.mu.RUnlock()
	}

	result, err := fn(client)
	if err != nil {
		return zero, err
	}

	return result, nil
}
