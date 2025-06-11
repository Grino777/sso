package keys

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Grino777/sso/internal/services/jwks/models"
	"github.com/google/uuid"
)

const opKeys = "keys."

var (
	errKeyNotExist = errors.New("private key not exist")
)

type Keys struct {
	PrivateKey *models.PrivateKey
	PublicKey  *models.PublicKey
}

// Работает только с файлами ключей внутри keysDir
type KeysManager struct {
	log     *slog.Logger
	keysDir string
}

func NewKeysManager(
	log *slog.Logger,
	keysDir string,
) *KeysManager {
	return &KeysManager{
		log:     log,
		keysDir: keysDir,
	}
}

// ----------------------------------Init Block---------------------------------

// Загружает и возвращает существувющие ключи из keysDir
func (km *KeysManager) LoadKeys() (*Keys, error) {
	const op = opKeys + "LoadKeys"

	if err := km.checkKeysFolder(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	privateKey, err := km.GetLatestPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	publicKey, err := km.SavePublicKeyToFile(privateKey)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	keys := &Keys{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}
	return keys, nil
}

// ----------------------------------End Block----------------------------------

// ------------------------------Private Key Block------------------------------

// func (km *KeysManager) GetPrivateKey() (*models.PrivateKey, error) {
// 	const op = opKeys + "retrievePrivateKey"

// 	privateKey, err := getLatestPrivateKey(km)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err := deletePublicKeysFiles(km); err != nil {
// 		return nil, fmt.Errorf("%s: %w", op, err)
// 	}
// 	km.log.Debug("private key successfully received", slog.String("op", op))
// 	return privateKey, nil
// }

func (km *KeysManager) GetLatestPrivateKey() (*models.PrivateKey, error) {
	const op = opKeys + "getLatestPrivateKey"

	pemFiles, err := km.listPrivateKeysFiles(km.keysDir)
	if errors.Is(err, errKeyNotExist) {
		privateKey, err := km.GeneratePrivateKey()
		if err != nil {
			return nil, err
		}
		return privateKey, nil
	} else if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	latestKey, err := km.getLatestPrivateKey(pemFiles)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := km.deleteOldPrivateKeys(pemFiles); err != nil {
		return nil, err
	}

	return latestKey, nil
}

func (ks *KeysManager) deleteOldPrivateKeys(pemFiles []os.DirEntry) error {
	const op = opKeys + "deleteOldPrivateKeys"

	for _, entry := range pemFiles[1:] {
		keyID := filepath.Base(entry.Name())
		if ext := filepath.Ext(keyID); ext != "" {
			keyID = keyID[:len(keyID)-len(ext)]
		}
		ks.mu.Lock()
		delete(ks.PrivateKeys, keyID)
		defer ks.mu.Unlock()
		oldKeyPath := filepath.Join(ks.KeysDir, entry.Name())
		if err := os.Remove(oldKeyPath); err != nil {
			return fmt.Errorf("%s: failed to remove old key %s: %w", op, oldKeyPath, err)
		}
	}
	return nil
}

func (km *KeysManager) getLatestPrivateKey(pemFiles []os.DirEntry) (*models.PrivateKey, error) {
	const op = opKeys + "getLatestPrivateKey"

	sort.Slice(pemFiles, func(i, j int) bool {
		infoI, _ := pemFiles[i].Info()
		infoJ, _ := pemFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	latestKeyPath := filepath.Join(km.keysDir, pemFiles[0].Name())
	privatePEM, err := os.ReadFile(latestKeyPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read private key %s: %w", op, latestKeyPath, err)
	}
	privateBlock, _ := pem.Decode(privatePEM)
	if privateBlock == nil {
		return nil, fmt.Errorf("%s: failed to decode private key PEM", op)
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse private key: %w", op, err)
	}

	keyID := filepath.Base(pemFiles[0].Name())
	if ext := filepath.Ext(keyID); ext != "" {
		keyID = keyID[:len(keyID)-len(ext)]
	}

	return &models.PrivateKey{
		ID:      keyID,
		Key:     privateKey,
		IsSaved: true,
	}, nil
}

func (km *KeysManager) GeneratePrivateKey() (*models.PrivateKey, error) {
	const op = opKeys + "GenerateKeyPrivateKey"

	pk, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		return nil, fmt.Errorf("%s: private key not generated: %w", op, err)
	}

	keyID, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	pkObj := &models.PrivateKey{
		ID:  keyID.String(),
		Key: pk,
	}

	if err := km.savePrivateKeyToFile(pkObj); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pkObj, nil
}

// FIXME For KeyStore
// func (km *KeysManager) setPrivateKey(privateKey *models.PrivateKey) error {
// 	const op = opKeys + "setPrivateKey"

// 	ks.mu.Lock()
// 	defer ks.mu.Unlock()

// 	if len(ks.PrivateKeys) > 1 {
// 		for _, v := range ks.PrivateKeys {
// 			if err := deletePrivateKey(ks, v); err != nil {
// 				return fmt.Errorf("%s: %w", op, err)
// 			}
// 		}
// 	}

// 	ks.PrivateKeys[privateKey.ID] = privateKey

// 	if !privateKey.IsSaved {
// 		if err := savePrivateKeyToFile(ks, privateKey); err != nil {
// 			return fmt.Errorf("%s: %w", op, err)
// 		}
// 		privateKey.IsSaved = true
// 	}
// 	if err := savePublicKeyToFile(ks, privateKey); err != nil {
// 		return fmt.Errorf("%s: failed to save public key: %w", op, err)
// 	}

// 	return nil
// }

func (km *KeysManager) savePrivateKeyToFile(privateKey *models.PrivateKey) error {
	const op = opKeys + "savePrivateKeyToFile"

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey.Key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	filePath := filepath.Join(km.keysDir, privateKey.ID+".pem")
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("%s: failed to create key file: %w", op, err)
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		return fmt.Errorf("%s: failed to encode private key: %w", op, err)
	}
	return nil
}

func (km *KeysManager) DeletePrivateKey(privateKey *models.PrivateKey) error {
	const op = opKeys + "deletePrivateKey"

	if err := deletePrivateKeyFile(privateKey.ID, km.keysDir); err != nil {
		if os.IsNotExist(err) {
			km.log.Debug("private key file not found, skipping deletion", slog.String("op", op), slog.String("keyID", privateKey.ID))
		} else {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()
	delete(ks.PrivateKeys, privateKey.ID)

	km.log.Debug("private key %s is deleted", privateKey.ID, slog.String("op", op))

	return nil
}

func (km *KeysManager) DeletePrivateKeyFile(keyID string) error {
	const op = opKeys + "deletePrivateKeyFile"

	fileName := fmt.Sprintf("%s.pem", keyID)
	path := path.Join(km.keysDir, fileName)

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// ----------------------------------End Block----------------------------------

// -------------------------------Public Key Block------------------------------

// func getPublicKey(
// 	ks *KeysStore,
// 	privateKey *rsa.PrivateKey,
// ) error {
// 	const op = "jwks.keys.utils.getPublicKey"
// 	panic("implement me!")
// }

// Удаляет все все публичные ключи из keys folder
func (km *KeysManager) deletePublicKeysFiles() error {
	const op = opKeys + "deletePublicKeys"

	entries, err := os.ReadDir(km.keysDir)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var errs []error
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pub.pem") {
			if err := os.Remove(filepath.Join(km.keysDir, entry.Name())); err != nil {
				errs = append(errs, fmt.Errorf("%s: failed to remove public key %s: %w", op, entry.Name(), err))
			} else {
				keyID := filepath.Base(entry.Name())
				if ext := filepath.Ext(keyID); ext != "" {
					keyID = keyID[:len(keyID)-len(ext)]
				}
				delete(ks.PublicKeys, keyID)
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("%s: %w", op, errors.Join(errs...))
	}
	km.log.Info("public keys deleted", slog.String("op", op))
	return nil
}

func (km *KeysManager) SavePublicKeyToFile(privateKey *models.PrivateKey) (*models.PublicKey, error) {
	const op = opKeys + "savePublicKeyToFile"

	filePath := filepath.Join(km.keysDir, privateKey.ID+".pub.pem")

	pubASN1 := x509.MarshalPKCS1PublicKey(&privateKey.Key.PublicKey)
	pubPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to create key file: %w", op, err)
	}
	defer file.Close()

	if err := pem.Encode(file, pubPEM); err != nil {
		return nil, fmt.Errorf("%s: failed to write public key to file: %w", op, err)
	}

	publicKey := &models.PublicKey{
		ID:      privateKey.ID,
		Key:     &privateKey.Key.PublicKey,
		IsSaved: true,
	}
	return publicKey, nil
}

// ----------------------------------End Block----------------------------------

func (km *KeysManager) checkKeysFolder() error {
	const op = opKeys + "checkKeysFolder"

	if _, err := os.Stat(km.keysDir); os.IsNotExist(err) {
		if err := km.createKeysFolder(); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	} else if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

func (km *KeysManager) createKeysFolder() error {
	const op = opKeys + "createKeysFolder"

	if err := os.Mkdir(km.keysDir, 0700); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	km.log.Debug("keys folder created", slog.String("op", op))
	return nil
}

// Перевод PublicKey в формат для передачи клиентам
// func (km *KeysManager) convertPubKeyToJWKS(pubKeyObj *models.PublicKey) (*models.JWKSToken, error) {
// 	pubKey := pubKeyObj.Key
// 	n := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
// 	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes())

// 	return &models.JWKSToken{
// 		Kid: pubKeyObj.ID,
// 		Kty: "RSA",
// 		Alg: "RS256",
// 		Use: "sig",
// 		N:   n,
// 		E:   e,
// 	}, nil
// }

// FIXME for KeyStore
// gorutine для отложенного удаления public key через определенное время
// func (km *KeysManager) deletePublicKeyTask(ctx context.Context, id string) error {
// 	const op = opKeys + "deletePublicKeyTask"

// 	timer := time.NewTimer(ks.TokenTTL + time.Minute)
// 	defer timer.Stop()

// 	go func() {
// 		select {
// 		case <-timer.C:
// 			ks.mu.Lock()
// 			defer ks.mu.Unlock()
// 			if _, exists := ks.PublicKeys[id]; exists {
// 				delete(ks.PublicKeys, id)
// 				ks.Log.Info("public key deleted", slog.String("op", op), slog.String("keyID", id))
// 			} else {
// 				ks.Log.Debug("public key not found, skipping deletion", slog.String("op", op), slog.String("keyID", id))
// 			}
// 		case <-ctx.Done():
// 			ks.Log.Debug("public key deletion cancelled", slog.String("op", op), slog.String("keyID", id))
// 			return
// 		}
// 	}()

// 	return nil
// }

func (km *KeysManager) listPrivateKeysFiles(keysDir string) ([]os.DirEntry, error) {
	const op = opKeys + "listPrivateKeysFiles"

	entries, err := os.ReadDir(keysDir)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read directory: %w", op, err)
	}

	var pemFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && path.Ext(entry.Name()) == ".pem" && !strings.HasSuffix(entry.Name(), ".pub.pem") {
			pemFiles = append(pemFiles, entry)
		}
	}

	if len(pemFiles) == 0 {
		return nil, errKeyNotExist
	}
	return pemFiles, nil
}
