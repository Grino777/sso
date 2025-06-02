package jwks

import (
	"context"

	"github.com/Grino777/sso-proto/gen/go/sso"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JwksService interface {
	GetJwks(context.Context) ([]*sso.Jwk, error)
}

type JwksServer struct {
	sso.UnimplementedJwksServer
	jwks JwksService
}

func RegService(s *grpc.Server, jwks JwksService) {
	sso.RegisterJwksServer(s, &JwksServer{jwks: jwks})
}

func (j *JwksServer) GetJwks(
	ctx context.Context,
	req *sso.GetJwksRequest,
) (*sso.GetJwksResponse, error) {
	tokensList, err := j.jwks.GetJwks(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &sso.GetJwksResponse{Keys: tokensList}, nil
}
