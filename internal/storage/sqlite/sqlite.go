// Пакет для взаимодействия с БД. Обрабатывает "запросы" приходящие от бизнес-логики.
package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models"
	"github.com/Grino777/sso/internal/storage"

	"github.com/mattn/go-sqlite3"
)

var (
	ErrRefreshTokenExist = errors.New("refresh token is exist in refresh_tokens table")
)

const sqliteOp = "storage.sqlite."

type SQLiteStorage struct {
	db         *sql.DB
	driverName string
	localPath  string
	superuser  config.SuperUser
	logger     *slog.Logger
}

// Creates a new DB session and performs migrations
func New(
	driverName, localPath string,
	superuser config.SuperUser,
	log *slog.Logger,
) *SQLiteStorage {
	return &SQLiteStorage{
		driverName: driverName,
		localPath:  localPath,
		superuser:  superuser,
		logger:     log,
	}
}

func (s *SQLiteStorage) SaveUser(
	ctx context.Context,
	username, passHash string,
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
		return user, fmt.Errorf("%s: %w", op, err)
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
		return app, fmt.Errorf("%s: %w", op, err)
	}
	return app, nil
}

// FIXME
func (s *SQLiteStorage) IsAdmin(
	ctx context.Context,
	username string,
) (bool, error) {
	panic("implement me")
}

func (s *SQLiteStorage) DeleteRefreshToken(
	ctx context.Context,
	userID uint64,
	appID uint32,
	token models.Token,
) error {
	query := `
        DELETE FROM refresh_tokens WHERE user_id = ? AND app_id = ? AND r_token = ?
    `
	_, err := s.db.ExecContext(ctx, query, userID, appID, token.Token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}
	return nil
}

func (s *SQLiteStorage) SaveRefreshToken(
	ctx context.Context,
	userID uint64,
	appID uint32,
	token models.Token,
) error {
	const op = "storage.sqlite.sqlite.SaveRefreshToken"

	query := `
		INSERT INTO refresh_tokens (user_id, app_id, r_token, expire_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (user_id, app_id) DO UPDATE
		SET r_token = excluded.r_token, expire_at = excluded.expire_at
	`

	_, err := s.db.ExecContext(ctx, query, userID, appID, token.Token, token.Expire_at)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.Code == sqlite3.ErrConstraint {
			if strings.Contains(sqliteErr.Error(), "refresh_tokens.r_token") {
				return fmt.Errorf("%s: %w", op, ErrRefreshTokenExist)
			}
		}
		return fmt.Errorf("failed to save or update refresh token: %w", err)
	}
	return nil
}
