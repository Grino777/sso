package keys

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Grino777/sso/internal/services/jwks/models"
)

type KeysStore struct {
	Log         *slog.Logger
	KeysDir     string
	PrivateKeys map[string]*models.PrivateKey
	PublicKeys  map[string]*models.PublicKey
	TokenTTL    time.Duration
}

func New(
	log *slog.Logger,
	keysDir string,
	tokenTTL time.Duration,
) (*KeysStore, error) {
	const op = "jwks.keys.New"

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

func (ks *KeysStore) GetPublicKeys() ([]map[string]any, error) {
	const op = "jwks.keys.GetPublicKeys"

	data := []map[string]any{}

	for _, v := range ks.PublicKeys {
		key, err := convertToJWKS(ks, v.Key)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", op, err)
		}
		data = append(data, key)
	}
	return data, nil
}

// Производит замену ключей (приватного и публичного)
// func (ks *KeysStore) RotateKeys() error {
// 	const op = "jwks.keys.RotateKeys"

// 	if ks.PrivateKeys == nil {
// 		if err := generatePrivateKey(ks); err != nil {
// 			return fmt.Errorf("%s: %v", op, err)
// 		}
// 	} else {
// 		if err := setPrivateKey(ks); err != nil {
// 			return fmt.Errorf("%s: %v", op, err)
// 		}
// 	}
// 	return nil
// }
