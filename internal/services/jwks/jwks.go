package jwks

import (
	"context"
	"log/slog"

	"github.com/Grino777/sso-proto/gen/go/sso"
	"github.com/Grino777/sso/internal/services/keys/models"
)

type KeysStore interface {
	GetPublicKeys() ([]*models.JWKSToken, error)
}

type JwksService struct {
	sso.UnimplementedJwksServer
	log       *slog.Logger
	keysStore KeysStore
}

func NewJwksService(
	log *slog.Logger,
	keysStore KeysStore,
) (*JwksService, error) {
	const op = "jwks.New"

	return &JwksService{
		log:       log,
		keysStore: keysStore,
	}, nil
}

func (j *JwksService) GetJwks(context.Context) ([]*sso.Jwk, error) {
	const op = "jwks.jwks.GetJwks"

	publicKeys, err := j.keysStore.GetPublicKeys()
	if err != nil {
		j.log.Error("%s: %w", op, err)
		return nil, err
	}

	data := []*sso.Jwk{}
	for _, token := range publicKeys {
		convertedToken := token.ConvertToken()
		data = append(data, convertedToken)
	}

	return data, nil
}

// func (j *JwksService) GetLatestPrivateKey(ctx context.Context) (*models.PrivateKey, error) {
// 	privateKey := j.keysStore.GetLatestPrivateKey()
// 	return privateKey, nil
// }
