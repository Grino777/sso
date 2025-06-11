package keys

import (
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Grino777/sso/internal/services/jwks/models"
)

const (
	keysOp = "jwks.keys.keys."
)

type KeysStore struct {
	Log         *slog.Logger
	KeysDir     string
	PrivateKeys map[string]*models.PrivateKey
	PublicKeys  map[string]*models.PublicKey
	TokenTTL    time.Duration
	mu          sync.RWMutex
}

func NewKeysStore(
	log *slog.Logger,
	keysDir string,
	tokenTTL time.Duration,
) (*KeysStore, error) {
	const op = keysOp + "New"

	ks := &KeysStore{
		Log:         log,
		KeysDir:     keysDir,
		TokenTTL:    tokenTTL,
		PrivateKeys: make(map[string]*models.PrivateKey),
		PublicKeys:  make(map[string]*models.PublicKey),
	}

	if err := initKeys(ks); err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	return ks, nil
}

func (ks *KeysStore) GetPublicKeys() ([]*models.JWKSToken, error) {
	const op = keysOp + "GetPublicKeys"

	data := []*models.JWKSToken{}

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	for _, v := range ks.PublicKeys {
		key, err := convertPubKeyToJWKS(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", op, err)
		}
		data = append(data, key)
	}
	return data, nil
}

func (ks *KeysStore) GetLatestPrivateKey() (*models.PrivateKey, error) {
	privateKey, err := getLatestPrivateKey(ks)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
