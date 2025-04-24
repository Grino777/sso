package sqlite

import "github.com/Grino777/sso/internal/storage"

func ToStorage(s *SQLiteStorage) storage.Storage {
	return storage.Storage{
		Users:  s,
		Apps:   s,
		Closer: s,
	}
}
