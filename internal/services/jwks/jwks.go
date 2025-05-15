package jwks

import (
	"context"
	"log/slog"

	"github.com/Grino777/sso-proto/gen/go/sso"
)

// type PublicKeys interface {
// 	AddPublicKeys() error
// 	DeletePublicKeys() error
// }

// type Keys interface {
// 	PublicKeys
// 	Load() error
// 	PublicKeys() error
// }

type KeysStore struct {
}

type JwksService struct {
	sso.UnimplementedJwksServer
	log  *slog.Logger
	keys KeysStore
}

func New(
	log *slog.Logger,
) *JwksService {
	panic("implement me!")
	// return &JwksService{
	// 	log:        log,
	// 	publicKeys: publicKeys,
	// }
}

// FIXME
func (j *JwksService) GetJwks(context.Context) error {
	panic("implement me!")
}
