package keys

import (
	"fmt"
	"log/slog"

	"github.com/Grino777/sso/internal/services/jwks/models"
)

type KeysStore struct {
	log        *slog.Logger
	keysDir    string
	PrivateKey *models.PrivateKey
	PublicKeys []*models.PublicKey
	// PublicKeys map[string]*rsa.PublicKey
}

func New(
	log *slog.Logger,
	keysPath string,
) (*KeysStore, error) {
	const op = "jwks.keys.New"

	ks := &KeysStore{
		keysDir: keysPath,
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
		key, err := toJWKS(ks, v.Key)
		if err != nil {
			return nil, fmt.Errorf("%s: %v", op, err)
		}
		data = append(data, key)
	}
	return data, nil
}

// Производит замену ключей (приватного и публичного)
func (ks *KeysStore) RotateKeys() error {
	const op = "jwks.keys.RotateKeys"

	if ks.PrivateKey == nil {
		if err := genPrivateKey(ks); err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
	} else {
		if err := setPrivateKey(ks); err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
	}
	return nil
}
