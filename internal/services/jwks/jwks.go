package jwks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Grino777/sso-proto/gen/go/sso"
	"github.com/Grino777/sso/internal/services/jwks/keys"
	"github.com/Grino777/sso/internal/services/jwks/models"
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

func (j *JwksService) GetJwks(context.Context) ([]*sso.Jwk, error) {
	const op = "jwks.jwks.GetJwks"

	publicKeys, err := j.keysStore.GetPublicKeys()
	if err != nil {
		j.log.Error("%s: %v", op, err)
		return nil, err
	}

	data := []*sso.Jwk{}
	for _, token := range publicKeys {
		convertedToken := token.ConvertToken()
		data = append(data, convertedToken)
	}

	return data, nil
}

func (j *JwksService) GetLatestPrivateKey(ctx context.Context) (*models.PrivateKey, error) {
	privateKey, err := j.keysStore.GetLatestPrivateKey()
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
