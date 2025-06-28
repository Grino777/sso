package manager

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/Grino777/sso/internal/services/keys/models"
)

const opKeysManager = "keys."

var (
	ErrKeyNotExist = errors.New("private key not exist")
	ErrDirNotFound = errors.New("keys dir not found")
	ErrKeyExpired  = errors.New("private key is expired")
)

type Keys struct {
	PrivateKey *models.PrivateKey
	PublicKeys map[string]*models.PublicKey
}

type GenKeys struct {
	PrivateKey *models.PrivateKey
	PublicKey  *models.PublicKey
}

// Работает только с файлами ключей внутри keysDir
type KeysManager struct {
	log      *slog.Logger
	keysDir  string
	keyTTL   time.Duration
	tokenTTL time.Duration
}

func NewKeysManager(
	log *slog.Logger,
	pathConfig config.PathConfig,
	ttlConfig config.TTLConfig,
) (*KeysManager, error) {
	km := &KeysManager{
		log:      log,
		keysDir:  pathConfig.KeysDir,
		keyTTL:   ttlConfig.KeyTTL,
		tokenTTL: ttlConfig.TokenTTL,
	}

	if err := km.initManager(); err != nil {
		return nil, err
	}
	return km, nil
}

func (km *KeysManager) LoadKeys() (*Keys, error) {
	const op = opKeysManager + "LoadKeys"

	keys := &Keys{
		PublicKeys: make(map[string]*models.PublicKey),
	}

	entries, err := os.ReadDir(km.keysDir)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	pemFiles := km.getPemFiles(entries)
	if len(pemFiles) == 0 {
		keys, err := km.generateNewPair()
		if err != nil {
			return nil, err
		}
		return keys, nil
	}
	uuids, err := sortPemFiles(pemFiles)
	if err != nil {
		return nil, err
	}

	activeKeys := km.decodePemFiles(uuids)
	if len(activeKeys) == 0 {
		keys, err := km.generateNewPair()
		if err != nil {
			return nil, err
		}
		return keys, nil
	}

	keys.PrivateKey = activeKeys[0]
	for _, key := range activeKeys {
		keys.PublicKeys[key.ID] = key.GetPublicKey()
	}
	return keys, nil
}

func (km *KeysManager) GeneratePairKeys() (*GenKeys, error) {
	const op = opKeysManager + "generatePrivateKey"

	km.log.Debug("generating new pair keys")

	privateKey, err := models.NewPrivateKey(km.keyTTL, km.tokenTTL, km.keysDir)
	if err != nil {
		return nil, err
	}

	if err := privateKey.SaveToFile(km.keysDir); err != nil {
		return nil, err
	}

	keys := &GenKeys{
		PrivateKey: privateKey,
		PublicKey:  privateKey.GetPublicKey(),
	}
	return keys, nil
}

// RotateKeys rotates the keys in the store.
func (km *KeysManager) RotateKeys() (*GenKeys, error) {
	const op = opKeysManager + "RotateKeys"

	keys, err := km.GeneratePairKeys()
	if err != nil {
		return nil, err
	}

	km.log.Debug("successfully rotate keys")
	return keys, nil
}

// Initializes the key manager
func (km *KeysManager) initManager() error {
	const op = opKeysManager + "initManager"

	if err := checkKeysFolder(km.keysDir); err != nil {
		if errors.Is(err, ErrDirNotFound) {
			if err := createKeysFolder(km.keysDir); err != nil {
				return err
			}
			km.log.Debug("keys folder created", slog.String("op", op))
		} else {
			return err
		}
	}
	return nil
}

func (km *KeysManager) getPemFiles(entries []os.DirEntry) []os.DirEntry {
	var pemFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && path.Ext(entry.Name()) == ".pem" {
			pemFiles = append(pemFiles, entry)
		}
	}
	return pemFiles
}

func (km *KeysManager) decodePemFiles(sortedUuids []os.DirEntry) []*models.PrivateKey {
	const op = opKeysManager + "decodePemFiles"

	activeKeys := make([]*models.PrivateKey, 0, len(sortedUuids))
	for _, file := range sortedUuids {
		filename := file.Name()
		privateKey, err := models.DecodePemToPrivateKey(file.Name(), km.keysDir, km.keyTTL, km.tokenTTL)
		if err != nil {
			if errors.Is(err, models.ErrPublicKeyExpired) {
				km.log.Debug("pem file is expired", "filename", filename)
			} else {
				km.log.Error("failed to decode pem file", "filename", filename, logger.Error(err))
			}
			if err := km.deletePemFile(filename); err != nil {
				km.log.Error("%s: failed to remove pem file %v", filename, err)
				continue
			}
		}
		activeKeys = append(activeKeys, privateKey)
	}
	return activeKeys
}

func (km *KeysManager) deletePemFile(pemUUID string) error {
	const op = opKeysManager + "deletePemFile"

	path := filepath.Join(km.keysDir, pemUUID)
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

func (km *KeysManager) generateNewPair() (*Keys, error) {
	keys := &Keys{
		PublicKeys: make(map[string]*models.PublicKey),
	}

	newKeys, err := km.GeneratePairKeys()
	if err != nil {
		return nil, err
	}
	keys.PrivateKey = newKeys.PrivateKey
	keys.PublicKeys[newKeys.PrivateKey.ID] = newKeys.PublicKey
	return keys, nil
}
