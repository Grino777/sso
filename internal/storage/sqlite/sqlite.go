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
func New(driverName string, dbPath string, dbUser config.DBUser) *SQLiteStorage {
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	s := SQLiteStorage{
		driverName: driverName,
		db:         db,
	}

	migrations.Migrate(s.db, driverName)
	sUtils.CreateSuperUser(s.db, dbUser.User, dbUser.Password)

	return &s
}

// Close DB session
func (s *SQLiteStorage) Close() error {
	s.db.Close()
	return nil
}

func (s *SQLiteStorage) SaveUser(
	ctx context.Context,
	username string,
	passHash string,
	// appID uint32,
) error {
	const op = "storage.SaveUser"

	stmt, err := s.db.Prepare("INSERT INTO users(username, pass_hash) VALUES(?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.ExecContext(ctx, username, passHash)
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
	appID uint32,
) (models.User, error) {
	const op = "storage.sqlite.GetUser"

	var user models.User

	query := "SELECT * FROM users WHERE username = ?"
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PassHash,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.User{}, storage.ErrUserNotFound
		}
		return models.User{}, fmt.Errorf("%s: %v", op, err)
	}
	return user, nil
}

func (s *SQLiteStorage) GetApp(
	ctx context.Context,
	appID uint32,
) (app models.App, err error) {
	query := "SELECT * FROM apps WHERE id = ?"
	err = s.db.QueryRowContext(ctx, query, appID).Scan(
		&app.ID,
		&app.Name,
		&app.Secret,
	)
	if err != nil {
		return models.App{}, err
	}

	return app, nil
}

// FIXME
func (s *SQLiteStorage) IsAdmin(
	ctx context.Context,
	username string,
) (isAdmin bool, err error) {
	panic("implement me")
}
