package jwks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Grino777/sso-proto/gen/go/sso"
	"github.com/Grino777/sso/internal/services/jwks/keys"
)

type JwksService struct {
	sso.UnimplementedJwksServer
	log       *slog.Logger
	keysStore *keys.KeysStore
}

func New(
	log *slog.Logger,
	keysDir string,
	tokenTTL time.Duration,
) (*JwksService, error) {
	const op = "jwks.New"

	ks, err := keys.New(log, keysDir, tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	return &JwksService{
		log:       log,
		keysStore: ks,
	}, nil
}

// FIXME
func (j *JwksService) GetJwks(context.Context) error {
	panic("implement me!")
}
