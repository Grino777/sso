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
	opStore = "keys.store."
)

// KeysStore represents the keys store
type KeysStore struct {
	mu          sync.RWMutex
	log         *slog.Logger
	PrivateKey  *models.PrivateKey
	PublicKeys  map[string]*models.PublicKey
	keysManager *manager.KeysManager
}

// NewKeysStore creates a new instance of KeysStore
func NewKeysStore(
	log *slog.Logger,
	pathConfig config.PathConfig,
	ttlConfig config.TTLConfig,
) (*KeysStore, error) {
	const op = opStore + "New"

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

// GetLatestPrivateKey returns the latest saved key
func (ks *KeysStore) GetLatestPrivateKey() (*models.PrivateKey, error) {
	const op = opStore + "GetLatestPrivateKey"

	log := logger.AddAttrs(ks.log, "op", op)

	ks.mu.Lock()
	defer ks.mu.Unlock()

	privateKey := ks.PrivateKey
	if privateKey == nil {
		log.Error("private key is nil")
		return nil, fmt.Errorf("private key is nil")
	}
	if privateKey.IsExpired() {
		newPair, err := ks.keysManager.GeneratePairKeys()
		if err != nil {
			return nil, err
		}
		ks.changePrivateKey(privateKey, newPair.PrivateKey)
		privateKey = newPair.PrivateKey
	}
	return privateKey, nil
}

// GenerateNewKeys generates a new pair of keys
func (ks *KeysStore) GenerateNewKeys() (*models.PrivateKey, error) {
	_, err := ks.keysManager.GeneratePairKeys()
	if err != nil {
		ks.log.Error("failed to generate new pair keys", logger.Error(err))
		return nil, err
	}
	newKey, err := ks.GetLatestPrivateKey()
	if err != nil {
		return nil, err
	}
	return newKey, nil
}

// RotateKeys generates a new pair of keys
func (ks *KeysStore) RotateKeys() (*manager.GenKeys, error) {
	const op = opStore + "RotateKeys"

	oldKeys := ks.PrivateKey

	keys, err := ks.keysManager.RotateKeys()
	if err != nil {
		return nil, err
	}
	ks.changePrivateKey(oldKeys, keys.PrivateKey)
	return keys, nil
}

func (ks *KeysStore) changePrivateKey(oldKey, newKey *models.PrivateKey) {
	const op = opStore + "removeExpiredKeys"

	ks.PrivateKey = newKey
	ks.PublicKeys[newKey.ID] = newKey.GetPublicKey()
	ks.PublicKeys[oldKey.ID] = oldKey.GetPublicKey()

	ks.log.Debug("%s: changed private key to %v", op, newKey.ID)
}

// GetPublicKeys returns a list of public keys in JWKS format.
// It removes expired keys and generates a new pair of keys if the current ones are missing.
// Returns an error if the operation fails.
func (ks *KeysStore) GetPublicKeys() ([]*models.JWKSToken, error) {
	const op = opStore + "GetPublicKeys"

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

// EnsurePublicKeys ensures that at least one public key is present.
// If keys are missing, it generates a new pair.
func (ks *KeysStore) EnsurePublicKeys() error {
	const op = opStore + "EnsurePublicKeys"

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

// removeExpiredKeys removes expired public keys.
func (ks *KeysStore) removeExpiredKeys() error {
	const op = opStore + "RemoveExpiredKeys"

	for _, publicKey := range ks.PublicKeys {
		if publicKey.IsExpired() {
			ks.log.Debug("%s: removing expired key with ID %s", op, publicKey.ID)
			if err := publicKey.DeletePair(); err != nil {
				return fmt.Errorf("%s: failed to delete expired key pair: %w", op, err)
			}
			delete(ks.PublicKeys, publicKey.ID)
		}
	}
	return nil
}
