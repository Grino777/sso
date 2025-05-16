package keys

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Grino777/sso/internal/services/jwks/models"
)

type KeysStore struct {
	Log         *slog.Logger
	KeysDir     string
	PrivateKeys map[string]*models.PrivateKey
	PublicKeys  map[string]*models.PublicKey
	TokenTTL    time.Duration
	mu          sync.RWMutex
}

func New(
	log *slog.Logger,
	keysDir string,
	tokenTTL time.Duration,
) (*KeysStore, error) {
	const op = "jwks.keys.keys.New"

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
	const op = "jwks.keys.keys.GetPublicKeys"

	data := []*models.JWKSToken{}

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	for _, v := range ks.PublicKeys {
		key, err := convertToJWKS(v)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", op, err)
		}
		data = append(data, key)
	}
	return data, nil
}

// Производит замену ключей (приватного и публичного)
func (ks *KeysStore) RotateKeys() error {
	const op = "jwks.keys.keys.RotateKeys"

	ctx := context.Background()

	oldPrivateKey, err := getLatestPrivateKey(ks)
	if err != nil {
		return err
	}
	if err := deletePrivateKey(ks, oldPrivateKey); err != nil {
		return err
	}
	newPrivateKey, err := generatePrivateKey(ks)
	if err != nil {
		return err
	}
	if err := setPrivateKey(ks, newPrivateKey); err != nil {
		return err
	}
	if err := deletePublicKeyTask(ctx, ks, oldPrivateKey.ID); err != nil {
		return err
	}

	ks.Log.Info("keys rotated successfully", slog.String("op", op), slog.String("newKeyID", newPrivateKey.ID))
	return nil
}
