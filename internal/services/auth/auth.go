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
	"github.com/Grino777/sso/internal/lib/jwt"
	"github.com/Grino777/sso/internal/lib/logger"
	jwksM "github.com/Grino777/sso/internal/services/jwks/models"
	"github.com/Grino777/sso/internal/storage"
	authU "github.com/Grino777/sso/internal/utils/auth"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// ---------------------------------DB Interfaces----------------------------------

type Storage interface {
	StorageUserProvider
	StorageAppProvider
	StorageTokenProvider
	Closer
}

type StorageUserProvider interface {
	SaveUser(ctx context.Context, user models.User, passHash string) error
	GetUser(ctx context.Context, username string) (models.User, error)
	IsAdmin(ctx context.Context, user models.User) (bool, error)
}

type StorageAppProvider interface {
	GetApp(ctx context.Context, appID uint32) (models.App, error)
}

type StorageTokenProvider interface {
	DeleteRefreshToken(userID uint64, appID uint32, token string) error
	SaveRefreshToken(userID uint64, appID uint32, token string) error
}

type Closer interface {
	Close() error
}

// -----------------------------------End Block------------------------------------

// --------------------------------Cache Interfaces--------------------------------

type CacheStorage interface {
	CacheUserProvider
	CacheAppProvider
	Closer
}

type CacheUserProvider interface {
	GetUser(ctx context.Context, username string, appID uint32) (models.User, error)
	SaveUser(ctx context.Context, user models.User, appID uint32) (models.User, error)
	IsAdmin(ctx context.Context, user models.User, app models.App) (bool, error)
}

type CacheAppProvider interface {
	GetApp(ctx context.Context, appID uint32) (models.App, error)
	SaveApp(ctx context.Context, app models.App) error
}

type JwksProvider interface {
	GetLatestPrivateKey(ctx context.Context) (*jwksM.PrivateKey, error)
}

// -----------------------------------End Block------------------------------------

// Объект для взаимодейсвтия с БД
type AuthService struct {
	Log             *slog.Logger
	DB              Storage
	Cache           CacheStorage
	TokenTTL        time.Duration
	RefreshTokenTTL time.Duration
	JwksService     JwksProvider
}

func New(
	authConfigs AuthService,
) *AuthService {
	return &AuthService{
		Log:             authConfigs.Log,
		DB:              authConfigs.DB,
		Cache:           authConfigs.Cache,
		TokenTTL:        authConfigs.TokenTTL,
		RefreshTokenTTL: authConfigs.RefreshTokenTTL,
		JwksService:     authConfigs.JwksService,
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

	user, err := s.getCachedUser(ctx, username, appID)
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

	app, err := s.getCachedApp(ctx, appID)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, fmt.Errorf("%s: %w", op, err)
	}

	privateKey, err := s.JwksService.GetLatestPrivateKey(ctx)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, err
	}

	tokens, err := jwt.CreateNewTokens(user, app, privateKey, s.TokenTTL, s.RefreshTokenTTL)
	if err != nil {
		log.Error("%s: %w", op, err)
		return models.Tokens{}, err
	}

	if err := s.DB.SaveRefreshToken(user.ID, appID, tokens.RefreshToken.Token); err != nil {
		if errors.Is()
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

	passHash, err := authU.CreatePassHash(user.Password)
	if err != nil {
		log.Error("failed to generate pass hash %v", logger.Error(err))
		return fmt.Errorf("%s: %v", op, err)
	}

	err = s.DB.SaveUser(ctx, user, passHash)
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
