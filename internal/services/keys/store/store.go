package store

import (
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

func (ks *KeysStore) GetPublicKeys() ([]*models.JWKSToken, error) {
	const op = keysOp + "GetPublicKeys"

	data := []*models.JWKSToken{}

	ks.mu.RLock()
	defer ks.mu.RUnlock()

	for _, publicKey := range ks.PublicKeys {
		if publicKey.CheckExpirationTime() {
			delete(ks.PublicKeys, publicKey.ID)
			if err := publicKey.DeletePair(); err != nil {
				return nil, err
			}
			continue
		}

		jwksToken := publicKey.ConvertToJWKS()
		data = append(data, jwksToken)
	}
	if len(ks.PublicKeys) == 0 {
		keys, err := ks.keysManager.GeneratePairKeys()
		if err != nil {
			return nil, err
		}
		ks.PrivateKey = keys.PrivateKey
		ks.PublicKeys[keys.PrivateKey.ID] = keys.PublicKey
		jwksToken := keys.PublicKey.ConvertToJWKS()
		data = append(data, jwksToken)
	}

	return data, nil
}

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
