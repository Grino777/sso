package jwks

import (
	"context"

	"github.com/Grino777/sso-proto/gen/go/sso"
	"google.golang.org/grpc"
)

type JwksService interface {
	GetJwks(context.Context) error
}

type JwksServer struct {
	sso.UnimplementedJwksServer
	jwks JwksService
}

func RegService(s *grpc.Server, jwks JwksService) {
	sso.RegisterJwksServer(s, &JwksServer{jwks: jwks})
}

// FIXME
func (j *JwksServer) GetJwks(
	ctx context.Context,
	req *sso.GetJwksRequest,
) (*sso.GetJwksResponse, error) {
	panic("implement me!")
}
