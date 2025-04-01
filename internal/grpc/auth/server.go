//Interface and methods for gRPC App struct

package auth

import (
	"context"

	sso_v1 "github.com/Grino777/sso-proto/gen/go/sso"
	"google.golang.org/grpc"
)

type Server struct {
	sso_v1.UnimplementedAuthServer
	auth Auth
}

// Register service in gRPC server
func Register(s *grpc.Server, auth Auth) {
	sso_v1.RegisterAuthServer(s, &Server{auth: auth})
}

type Auth interface {
	Login(ctx context.Context, username string, password string) (token string, err error)
	Logout(ctx context.Context, token string) (success bool, err error)
	RegisterNewUser(ctx context.Context, username string, password string) (success bool, err error)
	IsAdmin(ctx context.Context, token string) (is_admin bool, err error)
}

func (s *Server) Login(ctx context.Context, req *sso_v1.LoginRequest) (*sso_v1.LoginResponse, error) {
	panic("implement me!")
}

func (s *Server) Logout(ctx context.Context, req *sso_v1.LogoutRequest) (*sso_v1.LogoutResponse, error) {
	panic("implement me!")
}

func (s *Server) RegisterNewUser(ctx context.Context, req *sso_v1.RegisterRequest) (*sso_v1.RegisterResponse, error) {
	panic("implement me!")
}

func (s *Server) IsAdmin(ctx context.Context, req *sso_v1.IsAdminRequest) (*sso_v1.IsAdminResponse, error) {
	panic("implement me!")
}
