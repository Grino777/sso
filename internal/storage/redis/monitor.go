package redis

import (
	"context"
	"log/slog"
	"time"

	"github.com/Grino777/sso/internal/lib/logger"
)

func MonitorRedisConnection(
	ctx context.Context,
	rs *RedisStorage,
	log *slog.Logger,
	errChan chan error,
) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Debug("redis monitoring stopped due to context cancellation")
			return
		case <-ticker.C:
			rs.Mu.Lock()
			client := rs.Client
			rs.Mu.Unlock()

			if client == nil {
				log.Error("redis client is nil, attempting to reconnect")
				if err := rs.Connect(); err != nil {
					log.Error("failed to reconnect to Redis", logger.Error(err))
					errChan <- err
					return
				}
				continue
			}

			pingCtx, cancel := context.WithTimeout(context.Background(), rs.RetryDelay)
			if err := client.Ping(pingCtx).Err(); err != nil {
				log.Warn("redis ping failed, attempting to reconnect", logger.Error(err))
				if err := rs.Connect(); err != nil {
					log.Error("failed to reconnect to Redis", logger.Error(err))
					errChan <- err
					cancel()
					return
				}
			}
			cancel()
		}
	}
}
