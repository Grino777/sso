package auth

import (
	"context"
	"log/slog"
	"time"
)

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

type UserSaver interface {
	SaveUser(ctx context.Context, username string, password string) (string, error)
}

type UserProvider interface {
	GetUser(ctx context.Context, username string) (int, error)
	IsAdmin(ctx context.Context, username string) (bool, error)
}

type AppProvider interface {
	GetApp(ctx context.Context, appID int) (int, error)
}

func (s *AuthService) Login(ctx context.Context, username string, password string) (token string, err error) {
	panic("implement me")
}
func (s *AuthService) Logout(ctx context.Context, token string) (success bool, err error) {
	panic("implement me")
}
func (s *AuthService) RegisterNewUser(ctx context.Context, username string, password string) (success bool, err error) {
	panic("implement me")
}
func (s *AuthService) IsAdmin(ctx context.Context, username string) (isAdmin bool, err error) {
	panic("implement me")
}
