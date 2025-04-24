package redis

import "github.com/Grino777/sso/internal/storage"

func ToStorage(s *RedisStorage) storage.Storage {
	return storage.Storage{
		Users:  s,
		Apps:   s,
		Closer: s,
	}
}
