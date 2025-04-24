// Пакет для бизнес-логики. Обработка запросов с верхнего уровня (от клиента)
package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Grino777/sso/internal/domain/models"
	cjwt "github.com/Grino777/sso/internal/lib/jwt"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/Grino777/sso/internal/storage"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Объект для взаимодейсвтия с БД
type AuthService struct {
	log      *slog.Logger
	db       storage.Storage
	cache    storage.Storage
	tokenTTL time.Duration
}

func New(
	log *slog.Logger,
	db storage.Storage,
	cache storage.Storage,
	tokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		log:      log,
		db:       db,
		cache:    cache,
		tokenTTL: tokenTTL,
	}
}

func (s *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
	appID uint32,
) (token string, err error) {
	const op = "services.auth.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("username", username),
	)

	user, err := s.db.Users.GetUser(ctx, username, appID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			log.Error("user not found", logger.Error(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		log.Error("failed to get user", logger.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Error("invalid credentials", logger.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	app, err := s.db.Apps.GetApp(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	log.Info("logged is successfully", slog.String("username", username))

	tokenString, err := cjwt.NewToken(user, app, s.tokenTTL)
	if err != nil {
		return "", err
	}

	return tokenString, nil

}

func (s *AuthService) Register(
	ctx context.Context,
	username string,
	password string,
) error {
	const op = "services.auth.Register"

	log := s.log.With(
		slog.String("op", op),
		slog.String("username", username),
	)

	log.Info("registering user")

	passHash, err := createPassHash(password)
	if err != nil {
		log.Error("failed to generate pass hash", "user", username)

		return fmt.Errorf("%s: %v", op, err)
	}

	err = s.db.Users.SaveUser(ctx, username, passHash)
	if err != nil {
		log.Error("failed to save user", logger.Error(err))

		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// FIXME
func (s *AuthService) Logout(
	ctx context.Context,
	token string,
) (success bool, err error) {
	panic("implement me")
}

// FIXME
func (s *AuthService) IsAdmin(
	ctx context.Context,
	username string,
) (isAdmin bool, err error) {
	panic("implement me")
}

func (s *AuthService) GetApp(
	ctx context.Context,
	appID uint32,
) (models.App, error) {
	app, err := s.db.Apps.GetApp(ctx, appID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Error("record not found", "appID", appID)

			return models.App{}, err
		}
	}
	return app, nil
}

func createPassHash(pass string) (string, error) {
	const op = "services.auth.createPassHash"

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%s: %v", op, err)
	}
	return string(passHash), nil
}
