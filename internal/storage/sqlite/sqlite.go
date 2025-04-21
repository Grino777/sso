// Пакет для взаимодействия с БД. Обрабатывает "запросы" приходящие от бизнес-логики.
package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sso/internal/domain/models"
	"sso/internal/storage"
	"sso/migrations"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	driverName string
	db         *sql.DB
}

// Creates a new DB session and performs migrations
func New(driverName string, dbPath string) *Storage {
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	s := Storage{
		driverName: driverName,
		db:         db,
	}

	migrations.Migrate(s.db, driverName)

	return &s
}

// Close DB session
func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveUser(
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

func (s *Storage) GetUser(
	ctx context.Context,
	username string,
	appID uint32,
) (models.User, error) {
	panic("implemente me")
}

func (s *Storage) GetApp(
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

func (s *Storage) IsAdmin(
	ctx context.Context,
	username string,
) (isAdmin bool, err error) {
	panic("implement me")
}

// func saveUserToUserApps(ctx context.Context, userID int, appID int) error {
// 	panic("implement me")
// }
