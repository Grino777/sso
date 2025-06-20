package manager

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
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

	log := logger.AddAttrs(km.log, "op", op)

	keys := &Keys{
		PublicKeys: make(map[string]*models.PublicKey),
	}

	entries, err := os.ReadDir(km.keysDir)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	pemFiles := km.getPemFiles(entries)
	if len(pemFiles) == 0 {
		newKeys, err := km.GeneratePairKeys()
		if err != nil {
			return nil, err
		}

		keys.PrivateKey = newKeys.PrivateKey
		keys.PublicKeys[newKeys.PrivateKey.ID] = newKeys.PublicKey
		return keys, nil
	}

	uuids, err := sortPemFiles(pemFiles)
	if err != nil {
		return nil, err
	}

	activeKeys := make(map[int]*models.PrivateKey)

	for i, file := range uuids {
		privateKey, err := models.DecodePemToPrivateKey(file.Name(), km.keysDir, km.keyTTL, km.tokenTTL)
		if err != nil {
			if errors.Is(err, models.ErrPublicKeyExpired) {
				continue
			}
			log.Error("failed to decode pem file for %s key", file.Name(), logger.Error(err))
			continue
		}
		if i == 0 {
			keys.PrivateKey = privateKey
		}
		keys.PublicKeys[privateKey.ID] = privateKey.GetPublicKey()
		activeKeys[i] = privateKey
	}

	if err := deleteAllKeys(km.keysDir); err != nil {
		return nil, err
	}

	for _, key := range activeKeys {
		if key != nil {
			if err := key.SaveToFile(km.keysDir); err != nil {
				log.Error("failed to save private key: %s", key.ID, logger.Error(err))
				continue
			}
		}
	}

	return keys, nil
}

func (km *KeysManager) GeneratePairKeys() (*GenKeys, error) {
	const op = opKeysManager + "generatePrivateKey"

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
