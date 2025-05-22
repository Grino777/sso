// Пакет для взаимодействия с БД. Обрабатывает "запросы" приходящие от бизнес-логики.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models"
	"github.com/Grino777/sso/internal/storage"
	sUtils "github.com/Grino777/sso/internal/utils/storage/sqlite"
	"github.com/Grino777/sso/migrations"

	"github.com/mattn/go-sqlite3"
)

type SQLiteStorage struct {
	driverName string
	db         *sql.DB
}

// Creates a new DB session and performs migrations
func New(driverName string, dbPath string, dbUser config.DBUser) (*SQLiteStorage, error) {
	const op = "sqlite.New"

	storage := &SQLiteStorage{}

	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to connect to database: %v", op, err)
	}

	storage.driverName = driverName
	storage.db = db

	if err := migrations.Migrate(storage.db, driverName); err != nil {
		return nil, err
	}

	if err := sUtils.CreateSuperUser(storage.db, dbUser.User, dbUser.Password); err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	return storage, nil
}

// Close DB session
func (s *SQLiteStorage) Close() error {
	s.db.Close()
	return nil
}

func (s *SQLiteStorage) SaveUser(
	ctx context.Context,
	user *models.User,
	passHash string,
) error {
	const op = "storage.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users(username, pass_hash) VALUES(?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, user.Username, passHash)
	if err != nil {
		var sqlErr sqlite3.Error

		if errors.As(err, &sqlErr) && sqlErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return fmt.Errorf("%s: %w", op, storage.ErrUserExist)
		}

		return fmt.Errorf("%s: %w", op, err)
	}

	// userID, err := res.LastInsertId()
	// if err != nil {
	//
	// }

	return nil
}

func (s *SQLiteStorage) GetUser(
	ctx context.Context,
	username string,
) (models.User, error) {
	const op = "storage.sqlite.GetUser"

	user := models.User{}

	query := "SELECT * FROM users WHERE username = ?"
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PassHash,
		&user.Role_id,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return user, storage.ErrUserNotFound
		}
		return user, fmt.Errorf("%s: %v", op, err)
	}
	return user, nil
}

func (s *SQLiteStorage) GetApp(
	ctx context.Context,
	appID uint32,
) (app models.App, err error) {
	const op = "sqlite.GetApp"

	app = models.App{}

	query := "SELECT * FROM apps WHERE id = ?"
	err = s.db.QueryRowContext(ctx, query, appID).Scan(
		&app.ID,
		&app.Name,
		&app.Secret,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return app, storage.ErrUserNotFound
		}
		return app, fmt.Errorf("%s: %v", op, err)
	}

	return app, nil
}

// FIXME
func (s *SQLiteStorage) IsAdmin(
	ctx context.Context,
	user *models.User,
) (bool, error) {
	panic("implement me")
}

func (s *SQLiteStorage) DeleteRefreshToken(token string) error {
	panic("implement me!")
}

func (s *SQLiteStorage) SaveRefreshToken(token string) error {
	panic("implement me!")
}
