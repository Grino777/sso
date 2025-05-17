// Пакет для бизнес-логики. Обработка запросов с верхнего уровня (от клиента)
package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Grino777/sso/internal/domain/models"
	cjwt "github.com/Grino777/sso/internal/lib/jwt"
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

type DBStorage interface {
	DBUserProvider
	DBAppProvider
	Closer
}

type DBUserProvider interface {
	SaveUser(ctx context.Context, user *models.User, passHash string) error
	GetUser(ctx context.Context, username string) (*models.User, error)
	IsAdmin(ctx context.Context, user *models.User) (bool, error)
}

type DBAppProvider interface {
	GetApp(ctx context.Context, appID uint32) (*models.App, error)
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
	GetUser(ctx context.Context, user *models.User, app *models.App) (*models.User, error)
	SaveUser(ctx context.Context, user *models.User, app *models.App) error
	IsAdmin(ctx context.Context, user *models.User, app *models.App) (bool, error)
}

type CacheAppProvider interface {
	GetApp(ctx context.Context, appID uint32) (*models.App, error)
	SaveApp(ctx context.Context, app *models.App) error
}

type JwksProvider interface {
	GetLatestPrivateKey(ctx context.Context) (*jwksM.PrivateKey, error)
}

// -----------------------------------End Block------------------------------------

// Объект для взаимодейсвтия с БД
type AuthService struct {
	log         *slog.Logger
	db          DBStorage
	cache       CacheStorage
	tokenTTL    time.Duration
	jwksService JwksProvider
}

func New(
	log *slog.Logger,
	db DBStorage,
	cache CacheStorage,
	tokenTTL time.Duration,
	jwksService JwksProvider,
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
) (*models.Token, error) {
	const op = "services.auth.Login"

	log := s.log.With(
		slog.String("op", op),
		slog.String("username", username),
	)

	userObj, appObj, err := ValidateData(ctx, username, password, appID)
	if err != nil {
		log.Error("%s:%v", op, err)
		return nil, err
	}

	user, err := GetCachedUser(ctx, s.db, s.cache, userObj, appObj)
	if err != nil {
		log.Error("%s: %w", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Error("invalid credentials", logger.Error(err))
			return nil, ErrInvalidCredentials
		}
		log.Error("%s:%w", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	app, err := GetCachedApp(ctx, s.db, s.cache, appID)
	if err != nil {
		log.Error("%s: %w", op, err)
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	privateKey, err := s.jwksService.GetLatestPrivateKey(ctx)
	if err != nil {
		return nil, err
	}

	tokenObj, err := cjwt.NewAccessToken(user, app, privateKey, s.tokenTTL)
	if err != nil {
		log.Error("%s: %w", op, err)
		return nil, err
	}
	user.Token = tokenObj

	err = s.cache.SaveUser(ctx, user, app)
	if err != nil {
		log.Error("%s: %w", op, err)
		return nil, err
	}

	log.Info("logged is successfully")

	return tokenObj, nil

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

	_, err := s.db.GetUser(ctx, username)
	if err != nil {
		if !errors.Is(err, storage.ErrUserNotFound) {
			log.Error("%s:%v", op, err)
			return err
		}
	} else {
		log.Debug("user already exist")
		return storage.ErrUserExist
	}

	user, err := ValidateUser(username, password)
	if err != nil {
		log.Error("%s:%v", op, err)
		return err
	}

	passHash, err := authU.CreatePassHash(user.Password)
	if err != nil {
		log.Error("failed to generate pass hash %v", logger.Error(err))
		return fmt.Errorf("%s: %v", op, err)
	}

	err = s.db.SaveUser(ctx, user, passHash)
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
