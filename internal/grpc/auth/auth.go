// Пакет с методами для взаимодействия клиента и gRPC сервера
package auth

import (
	"context"
	"errors"
	"strconv"

	"github.com/Grino777/sso/internal/domain/models"
	"github.com/Grino777/sso/internal/services/auth"
	"github.com/Grino777/sso/internal/storage"

	sso_v1 "github.com/Grino777/sso-proto/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Методы для работы с бизнес-логикой
type AuthService interface {
	Login(ctx context.Context, username string, password string, appID uint32) (token *models.Token, err error)
	Logout(ctx context.Context, token string) (success bool, err error)
	Register(ctx context.Context, username string, password string) error
	IsAdmin(ctx context.Context, username string) (isAdmin bool, err error)
	RefreshToken(ctx context.Context, token, expired_at string) (*models.Token, error)
}

// Объект реализует gRPC-сервер для сервиса аутентификации с обязательными методами.
type AuthServer struct {
	sso_v1.UnimplementedAuthServer
	auth AuthService
}

// RegServer регистрирует gRPC-сервер аутентификации.
func RegServer(s *grpc.Server, auth AuthService) {
	sso_v1.RegisterAuthServer(s, &AuthServer{auth: auth})
}

func (s *AuthServer) Login(
	ctx context.Context,
	req *sso_v1.LoginRequest,
) (*sso_v1.LoginResponse, error) {
	token, err := s.auth.Login(ctx, req.GetUsername(), req.GetPassword(), req.Metadata.GetAppId())
	if err != nil {
		var valErr *models.ValidationError
		if errors.As(err, &valErr) {
			return nil, status.Error(codes.InvalidArgument, valErr.Error())
		}

		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid login or password")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &sso_v1.LoginResponse{
		AccessToken: &sso_v1.UserToken{
			Token:     token.Token,
			ExpiredAt: strconv.FormatInt(token.Expire_at, 10),
		}, // FIXME
	}, nil

}

// FIXME
func (s *AuthServer) Logout(
	ctx context.Context,
	req *sso_v1.LogoutRequest,
) (*sso_v1.LogoutResponse, error) {
	panic("implement me!")
}

func (s *AuthServer) Register(
	ctx context.Context,
	req *sso_v1.RegisterRequest,
) (*sso_v1.RegisterResponse, error) {
	err := s.auth.Register(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExist) {
			return nil, status.Error(codes.AlreadyExists, "user already exist")
		}

		var errVal *models.ValidationError
		if errors.As(err, &errVal) {
			return nil, status.Error(codes.InvalidArgument, errVal.Error())
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &sso_v1.RegisterResponse{Success: true}, nil
}

// FIXME
func (s *AuthServer) IsAdmin(
	ctx context.Context,
	req *sso_v1.IsAdminRequest,
) (*sso_v1.IsAdminResponse, error) {
	panic("implement me!")
}

func (s *AuthServer) RefreshToken(
	ctx context.Context,
	req *sso_v1.RefreshTokenRequest,
) (*sso_v1.LoginResponse, error) {
	panic("implement me!")
}
