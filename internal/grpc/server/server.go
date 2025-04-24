// Пакет с методами для взаимодействия клиента и gRPC сервера
package server

import (
	"context"
	"errors"
	"strings"

	"github.com/Grino777/sso/internal/services/auth"
	"github.com/Grino777/sso/internal/storage"

	sso_v1 "github.com/Grino777/sso-proto/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Методы для работы с бизнес-логикой
type AuthService interface {
	Login(ctx context.Context, username string, password string, appID uint32) (token string, err error)
	Logout(ctx context.Context, token string) (success bool, err error)
	Register(ctx context.Context, username string, password string) error
	IsAdmin(ctx context.Context, username string) (isAdmin bool, err error)
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
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.Metadata.AppId == 0 {
		return nil, status.Error(codes.InvalidArgument, "appID is required")
	}

	token, err := s.auth.Login(ctx, req.GetUsername(), req.GetPassword(), req.Metadata.GetAppId())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid login or password")
		}
	}

	return &sso_v1.LoginResponse{Token: token}, nil

}

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
	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if strings.Contains(req.Username, " ") {
		return nil, status.Error(codes.InvalidArgument, "username contains spaces")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	err := s.auth.Register(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		if errors.Is(err, storage.ErrUserExist) {
			return nil, status.Error(codes.AlreadyExists, "user already exist")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &sso_v1.RegisterResponse{Success: true}, nil
}

func (s *AuthServer) IsAdmin(
	ctx context.Context,
	req *sso_v1.IsAdminRequest,
) (*sso_v1.IsAdminResponse, error) {
	panic("implement me!")
}
