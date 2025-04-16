// Пакет для бизнес-логики. Обработка запросов с верхнего уровня (от клиента)
package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sso/internal/domain/models"
	cjwt "sso/internal/lib/jwt"
	"sso/internal/lib/logger"
	"sso/internal/storage"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// Объект для взаимодейсвтия с БД
type AuthService struct {
	log          *slog.Logger
	userSaver    UserSaver
	userProvider UserProvider
	appProvider  AppProvider
	tokenTTL     time.Duration
}

func New(
	log *slog.Logger,
	userSaver UserSaver,
	userProvider UserProvider,
	appProvider AppProvider,
	tokenTTL time.Duration,
) *AuthService {
	return &AuthService{
		log:          log,
		userSaver:    userSaver,
		userProvider: userProvider,
		appProvider:  appProvider,
		tokenTTL:     tokenTTL,
	}
}

// Логика сохранения пользователя
type UserSaver interface {
	SaveUser(ctx context.Context, username string, password string) error
}

// Логика получения пользователя и проверки на админа
type UserProvider interface {
	GetUser(ctx context.Context, username string, appID uint32) (models.User, error)
	IsAdmin(ctx context.Context, username string) (bool, error)
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

	log.Info("attempting to login user")

	user, err := s.userProvider.GetUser(ctx, username, appID)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			s.log.Warn("user not found", logger.Error(err))

			return "", fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		s.log.Error("failed to get user", logger.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		s.log.Error("invalid credentials", logger.Error(err))

		return "", fmt.Errorf("%s: %w", op, err)
	}

	app, err := s.appProvider.GetApp(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	s.log.Info("logged is successfully", slog.String("username", username))

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

	err := s.userSaver.SaveUser(ctx, username, password)
	if err != nil {
		log.Error("failed to save user", logger.Error(err))

		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *AuthService) Logout(
	ctx context.Context,
	token string,
) (success bool, err error) {
	panic("implement me")
}

func (s *AuthService) IsAdmin(
	ctx context.Context,
	username string,
) (isAdmin bool, err error) {
	panic("implement me")
}

// Логика получения приложения (в таблице БД)
type AppProvider interface {
	GetApp(ctx context.Context, appID uint32) (models.App, error)
}

func (s *AuthService) GetApp(
	ctx context.Context,
	appID uint32,
) (models.App, error) {
	app, err := s.appProvider.GetApp(ctx, appID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.log.Error("record not found", "appID", appID)

			return models.App{}, err
		}
	}

	return app, nil

}
