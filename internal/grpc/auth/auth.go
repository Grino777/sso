//Interface and methods for gRPC App struct

package auth

import (
	"context"

	sso_v1 "github.com/Grino777/sso-proto/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Auth - интерфейс для сервиса аутентификации.
type Auth interface {
	Login(ctx context.Context, username string, password string) (token string, err error)
	Logout(ctx context.Context, token string) (success bool, err error)
	RegisterNewUser(ctx context.Context, username string, password string) (success bool, err error)
	IsAdmin(ctx context.Context, username string) (isAdmin bool, err error)
}

// GRPCServer реализует gRPC-сервер для сервиса аутентификации.
type GRPCServer struct {
	sso_v1.UnimplementedAuthServer
	auth Auth
}

// Register регистрирует gRPC-сервер аутентификации.
func Register(s *grpc.Server, auth Auth) {
	sso_v1.RegisterAuthServer(s, &GRPCServer{auth: auth})
}

func (s *GRPCServer) Login(
	ctx context.Context,
	req *sso_v1.LoginRequest,
) (*sso_v1.LoginResponse, error) {

	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

	if req.Username == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}

}

func (s *GRPCServer) Logout(
	ctx context.Context,
	req *sso_v1.LogoutRequest,
) (*sso_v1.LogoutResponse, error) {
	panic("implement me!")
}

func (s *GRPCServer) RegisterNewUser(
	ctx context.Context,
	req *sso_v1.RegisterRequest,
) (*sso_v1.RegisterResponse, error) {
	panic("implement me!")
}

func (s *GRPCServer) IsAdmin(
	ctx context.Context,
	req *sso_v1.IsAdminRequest,
) (*sso_v1.IsAdminResponse, error) {
	panic("implement me!")
}
