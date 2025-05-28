// Пакет для бизнес-логики. Обработка запросов с верхнего уровня (от клиента)
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models"
	"github.com/Grino777/sso/internal/domain/models/interfaces"
	"github.com/Grino777/sso/internal/lib/logger"
	jwksM "github.com/Grino777/sso/internal/services/jwks/models"
	"github.com/Grino777/sso/internal/storage"
	authU "github.com/Grino777/sso/internal/utils/auth"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type JwksProvider interface {
	GetLatestPrivateKey(ctx context.Context) (*jwksM.PrivateKey, error)
}

// Объект для взаимодейсвтия с БД
type AuthService struct {
	Logger      *slog.Logger
	DB          interfaces.Storage
	Cache       interfaces.CacheStorage
	Tokens      config.TokenConfig
	JwksService JwksProvider
}

func New(
	authConfigs AuthService,
) *AuthService {
	return &AuthService{
		Logger:      authConfigs.Logger,
		DB:          authConfigs.DB,
		Cache:       authConfigs.Cache,
		Tokens:      authConfigs.Tokens,
		JwksService: authConfigs.JwksService,
	}
}

func (s *AuthService) Login(
	ctx context.Context,
	username string,
	password string,
	appID uint32,
) (models.Tokens, error) {
	const op = "services.auth.Login"

	log := s.Logger.With(
		slog.String("op", op),
		slog.String("username", username),
	)

	err := ValidateData(ctx, username, password, appID)
	if err != nil {
		log.Error("%s:%v", op, err)
		return models.Tokens{}, err
	}

	user, err := s.GetCachedUser(ctx, username, appID)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, err
	}

	if err := s.validatePassword(user.PassHash, password); err != nil {
		return models.Tokens{}, err
	}

	app, err := s.GetCachedApp(ctx, appID)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, err
	}

	user, err = s.generateUserTokens(ctx, user, app)
	if err != nil {
		return models.Tokens{}, err
	}

	_, err = s.Cache.SaveUser(ctx, user, appID)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, err
	}

	// TODO loging user login

	log.Info("logged is successfully")
	return user.Tokens, nil
}

func (s *AuthService) Register(
	ctx context.Context,
	username string,
	password string,
) error {
	const op = "services.auth.Register"

	log := s.Logger.With(
		slog.String("op", op),
		slog.String("username", username),
	)

	// First check if the user exists in the database.
	_, err := s.DB.GetUser(ctx, username)
	if err != nil {
		if !errors.Is(err, storage.ErrUserNotFound) {
			log.Error("%s:%v", op, err)
			return err
		}
	} else {
		log.Debug("user already exist")
		return storage.ErrUserExist
	}

	err = ValidateUser(username, password)
	if err != nil {
		log.Error("%s:%v", op, err)
		return err
	}

	passHash, err := authU.CreatePassHash(password)
	if err != nil {
		log.Error("failed to generate pass hash %v", logger.Error(err))
		return fmt.Errorf("%s: %v", op, err)
	}

	err = s.DB.SaveUser(ctx, username, passHash)
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

// FIXME
func (s *AuthService) RefreshToken(
	ctx context.Context,
	token, expired_at string,
) (*models.Token, error) {
	panic("implement me")
}

// func (s *AuthService) GetApp(
// 	ctx context.Context,
// 	appID uint32,
// ) (models.App, error) {
// 	app, err := s.db.GetApp(ctx, appID)
// 	if err != nil {
// 		if errors.Is(err, sql.ErrNoRows) {
// 			s.log.Error("record not found", "appID", appID)

// 			return models.App{}, err
// 		}
// 	}
// 	return app, nil
// }
