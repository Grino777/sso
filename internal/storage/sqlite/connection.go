package sqlite

import (
	"context"
	"database/sql"
	"fmt"

	sUtils "github.com/Grino777/sso/internal/utils/storage/sqlite"
	"github.com/Grino777/sso/migrations"
	_ "github.com/mattn/go-sqlite3"
)

func (s *SQLiteStorage) Connect(ctx context.Context) error {
	const op = sqliteOp + "Connect"

	db, err := sql.Open(s.driverName, s.localPath)
	if err != nil {
		return fmt.Errorf("%s: failed to connect to database: %v", op, err)
	}

	s.db = db

	if err := migrations.Migrate(s.db, s.driverName); err != nil {
		return err
	}
	if err := sUtils.CreateSuperUser(s.db, s.superuser.Username, s.superuser.Password); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	s.logger.Debug("database connection successfully")
	return nil
}

// Close DB session
func (s *SQLiteStorage) Close(ctx context.Context) error {
	const op = sqliteOp + "Close"

	if err := s.db.Close(); err != nil {
		s.logger.Error("failed to closing sqlite connection")
		s.db = nil
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}
