package redis

import (
	"sso/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	redis *redis.Client
}

func New(cfg config.RedisConfig) *RedisStore {
	options := &redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		Username:     cfg.User,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	}
	redis := redis.NewClient(options)
	return &RedisStore{
		redis: redis,
	}
}

func (r *RedisStore) Login() error {
	panic("implement me!")
}

func (r *RedisStore) Logout() error {
	panic("implement me!")
}
