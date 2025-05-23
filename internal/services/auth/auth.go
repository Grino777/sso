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
	"github.com/Grino777/sso/internal/lib/jwt"
	"github.com/Grino777/sso/internal/lib/logger"
	jwksM "github.com/Grino777/sso/internal/services/jwks/models"
	"github.com/Grino777/sso/internal/storage"
	"github.com/Grino777/sso/internal/storage/sqlite"
	authU "github.com/Grino777/sso/internal/utils/auth"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type JwksProvider interface {
	GetLatestPrivateKey(ctx context.Context) (*jwksM.PrivateKey, error)
}

// Объект для взаимодейсвтия с БД
type AuthService struct {
	Log         *slog.Logger
	DB          interfaces.Storage
	Cache       interfaces.CacheStorage
	Tokens      config.TokenConfig
	JwksService JwksProvider
}

func New(
	authConfigs AuthService,
) *AuthService {
	return &AuthService{
		Log:         authConfigs.Log,
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

	log := s.Log.With(
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
		return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Error("invalid credentials", logger.Error(err))
			return models.Tokens{}, ErrInvalidCredentials
		}
		log.Error("%s:%w", op, err)
		return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	app, err := s.GetCachedApp(ctx, appID)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	privateKey, err := s.JwksService.GetLatestPrivateKey(ctx)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, err
	}

	tokens, err := jwt.CreateNewTokens(user, app, privateKey, s.Tokens)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, err
	}

	for {
		if err := s.DB.SaveRefreshToken(ctx, user.ID, appID, tokens.RefreshToken); err != nil {
			if errors.Is(err, sqlite.ErrRefreshTokenExist) {
				s.Log.Debug("refresh token already exists, generating new token")
				refreshToken, err := jwt.NewRefreshToken(s.Tokens.RefreshTokenTTL)
				if err != nil {
					s.Log.Error("failed to generate new refresh token", logger.Error(err))
					return models.Tokens{}, fmt.Errorf("%s: failed to generate new refresh token: %w", op, err)
				}
				tokens.RefreshToken = refreshToken
			}
			s.Log.Debug("refresh token updated")
			break
		}
	}

	_, err = s.Cache.SaveUser(ctx, user, appID)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, err
	}

	log.Info("logged is successfully")

	return tokens, nil

}

func (s *AuthService) Register(
	ctx context.Context,
	username string,
	password string,
) error {
	const op = "services.auth.Register"

	log := s.Log.With(
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
