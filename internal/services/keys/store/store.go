package store

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/Grino777/sso/internal/services/keys/manager"
	"github.com/Grino777/sso/internal/services/keys/models"
)

const (
	keysOp = "jwks.keys.keys."
)

type KeysStore struct {
	mu          sync.RWMutex
	log         *slog.Logger
	PrivateKey  *models.PrivateKey
	PublicKeys  map[string]*models.PublicKey
	keysManager *manager.KeysManager
}

func NewKeysStore(
	log *slog.Logger,
	pathConfig config.PathConfig,
	ttlConfig config.TTLConfig,
) (*KeysStore, error) {
	const op = keysOp + "New"

	ks := &KeysStore{
		log:        log,
		PublicKeys: make(map[string]*models.PublicKey),
	}

	keysManager, err := manager.NewKeysManager(log, pathConfig, ttlConfig)
	if err != nil {
		return nil, err
	}

	keys, err := keysManager.LoadKeys()
	if err != nil {
		return nil, err
	}

	ks.PrivateKey = keys.PrivateKey
	ks.PublicKeys = keys.PublicKeys
	ks.keysManager = keysManager

	return ks, nil
}

// FIXME проверка на отсутствие паблик ключей в случае если они все протухшие
// func (ks *KeysStore) GetPublicKeys() ([]*models.JWKSToken, error) {
// 	const op = keysOp + "GetPublicKeys"

// 	data := []*models.JWKSToken{}

// 	ks.mu.RLock()
// 	defer ks.mu.RUnlock()

// 	for _, publicKey := range ks.PublicKeys {
// 		if publicKey.CheckExpirationTime() {
// 			delete(ks.PublicKeys, publicKey.ID)
// 			if err := publicKey.DeletePair(); err != nil {
// 				return nil, err
// 			}
// 			continue
// 		}

// 		jwksToken := publicKey.ConvertToJWKS()
// 		data = append(data, jwksToken)
// 	}
// 	if len(ks.PublicKeys) == 0 {
// 		keys, err := ks.keysManager.GeneratePairKeys()
// 		if err != nil {
// 			return nil, err
// 		}
// 		privateKey := keys.PrivateKey
// 		publicKey := privateKey.GetPublicKey()
// 		ks.PrivateKey = privateKey
// 		ks.PublicKeys[privateKey.ID] = publicKey
// 		jwksToken := keys.PublicKey.ConvertToJWKS()
// 		data = append(data, jwksToken)
// 	}

// 	return data, nil
// }

func (ks *KeysStore) GetLatestPrivateKey() *models.PrivateKey {
	return ks.PrivateKey
}

func (ks *KeysStore) GenerateNewKeys() (*models.PrivateKey, error) {
	_, err := ks.keysManager.GeneratePairKeys()
	if err != nil {
		ks.log.Error("failed to generate new pair keys", logger.Error(err))
		return nil, err
	}
	return ks.GetLatestPrivateKey(), nil
}

func (ks *KeysStore) RotateKeys() error {
	panic("implement me!")
}

// GetPublicKeys возвращает список публичных ключей в формате JWKS.
// Удаляет истёкшие ключи и генерирует новую пару ключей, если текущие отсутствуют.
// Возвращает ошибку, если операция не удалась.
func (ks *KeysStore) GetPublicKeys() ([]*models.JWKSToken, error) {
	const op = keysOp + "GetPublicKeys"

	ks.mu.Lock()
	defer ks.mu.Unlock()

	if err := ks.removeExpiredKeys(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := ks.EnsurePublicKeys(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	data := make([]*models.JWKSToken, 0, len(ks.PublicKeys))
	for _, publicKey := range ks.PublicKeys {
		data = append(data, publicKey.ConvertToJWKS())
	}

	return data, nil
}

// EnsurePublicKeys гарантирует наличие хотя бы одного публичного ключа.
// Если ключи отсутствуют, генерирует новую пару.
func (ks *KeysStore) EnsurePublicKeys() error {
	const op = keysOp + "EnsurePublicKeys"

	if ks.PublicKeys == nil {
		ks.PublicKeys = make(map[string]*models.PublicKey)
	}

	if len(ks.PublicKeys) > 0 {
		return nil
	}

	keys, err := ks.keysManager.GeneratePairKeys()
	if err != nil {
		return fmt.Errorf("%s: failed to generate key pair: %w", op, err)
	}

	ks.log.Debug("%s: generated new key pair with ID %s", op, keys.PrivateKey.ID)
	privateKey := keys.PrivateKey
	publicKey := privateKey.GetPublicKey()
	ks.PrivateKey = privateKey
	ks.PublicKeys[privateKey.ID] = publicKey
	return nil
}

// removeExpiredKeys удаляет истёкшие публичные ключи.
func (ks *KeysStore) removeExpiredKeys() error {
	const op = keysOp + "RemoveExpiredKeys"

	for _, publicKey := range ks.PublicKeys {
		if publicKey.CheckExpirationTime() {
			ks.log.Debug("%s: removing expired key with ID %s", op, publicKey.ID)
			if err := publicKey.DeletePair(); err != nil {
				return fmt.Errorf("%s: failed to delete expired key pair: %w", op, err)
			}
			delete(ks.PublicKeys, publicKey.ID)
		}
	}
	return nil
}
