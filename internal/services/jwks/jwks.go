package jwks

import (
	"context"
	"crypto/rsa"
	"log/slog"

	"github.com/Grino777/sso-proto/gen/go/sso"
)

type JwksService struct {
	sso.UnimplementedJwksServer
	log        *slog.Logger
	publicKeys map[string]*rsa.PublicKey
}

func New(
	log *slog.Logger,
	publicKeys map[string]*rsa.PublicKey,
) *JwksService {
	return &JwksService{
		log:        log,
		publicKeys: publicKeys,
	}
}

// FIXME
func (j *JwksService) GetJwks(context.Context) error {
	panic("implement me!")
}
