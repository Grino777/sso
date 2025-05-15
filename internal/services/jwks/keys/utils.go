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
	"time"

	"github.com/Grino777/sso/internal/services/jwks/models"
	"github.com/google/uuid"
)

var (
	errKeyNotExist = errors.New("private key not exist")
)

// ----------------------------------Init Block---------------------------------

func initKeys(ks *KeysStore) error {
	const op = "jwks.keys.initKeys"

	if err := loadKeys(ks); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

func loadKeys(ks *KeysStore) error {
	const op = "jwks.keys.loadKeys"

	if err := checkKeysFolder(ks); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	privateKey, err := retrievePrivateKey(ks)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	if err := setPrivateKey(ks, privateKey); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}

// ----------------------------------End Block----------------------------------

// ------------------------------Private Key Block------------------------------

func retrievePrivateKey(ks *KeysStore) (*models.PrivateKey, error) {
	const op = "jwks.keys.getPrivateKey"

	var privateKey *models.PrivateKey

	privateKey, err := getLatestPrivateKey(ks)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	if err := deletePublicKeys(ks); err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	ks.Log.Debug("private key successfully received", slog.String("op", op))
	return privateKey, nil
}

func getLatestPrivateKey(ks *KeysStore) (*models.PrivateKey, error) {
	const op = "utils.jwks.GetPrivateKey"

	pemFiles, err := listPrivateKeysFiles(ks.KeysDir)
	if errors.Is(err, errKeyNotExist) {
		privateKey, err := generatePrivateKey(ks)
		if err != nil {
			return nil, err
		}
		return privateKey, nil
	} else if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	latestKey, err := findLatestKey(ks, pemFiles)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", op, err)
	}

	return latestKey, nil
}

func findLatestKey(ks *KeysStore, pemFiles []os.DirEntry) (*models.PrivateKey, error) {
	const op = "jwks.keys.utils.latestKey"

	sort.Slice(pemFiles, func(i, j int) bool {
		infoI, _ := pemFiles[i].Info()
		infoJ, _ := pemFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	latestKeyPath := path.Join(ks.KeysDir, pemFiles[0].Name())
	privatePEM, err := os.ReadFile(latestKeyPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read private key %s: %v", op, latestKeyPath, err)
	}
	privateBlock, _ := pem.Decode(privatePEM)
	if privateBlock == nil {
		return nil, fmt.Errorf("%s: failed to decode private key PEM", op)
	}
	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse private key: %v", op, err)
	}

	keyID := filepath.Base(pemFiles[0].Name())
	if ext := filepath.Ext(keyID); ext != "" {
		keyID = keyID[:len(keyID)-len(ext)]
	}

	for _, entry := range pemFiles[1:] {
		oldKeyPath := path.Join(ks.KeysDir, entry.Name())
		if ext := filepath.Ext(keyID); ext != "" {
			keyID = keyID[:len(keyID)-len(ext)]
		}
		delete(ks.PrivateKeys, keyID)
		if err := os.Remove(oldKeyPath); err != nil {
			return nil, fmt.Errorf("%s: failed to remove old key %s: %v", op, oldKeyPath, err)
		}
	}

	return &models.PrivateKey{
		ID:  keyID,
		Key: privateKey,
	}, nil
}

func generatePrivateKey(ks *KeysStore) (*models.PrivateKey, error) {
	const op = "jwks.keys.utils.genPrivateKey"

	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("%s: private key not generated: %v", op, err)
	}

	keyID, err := uuid.NewV7()

	pkObj := &models.PrivateKey{
		ID:  keyID.String(),
		Key: pk,
	}

	return pkObj, nil
}

func setPrivateKey(ks *KeysStore, pk *models.PrivateKey) error {
	const op = "jwks.keys.utils.setPrivateKey"

	if len(ks.PrivateKeys) != 0 {
		for _, v := range ks.PrivateKeys {
			if err := deletePrivateKey(ks, v); err != nil {
				return fmt.Errorf("%s: %v", op, err)
			}
		}
	}

	ks.PrivateKeys[pk.ID] = pk

	ks.PublicKeys[pk.ID] = &models.PublicKey{
		ID:  pk.ID,
		Key: &pk.Key.PublicKey,
	}

	if err := savePrivateKeyToFile(pk, ks.KeysDir); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}

func savePrivateKeyToFile(pk *models.PrivateKey, kDir string) error {
	const op = "jwks.keys.utils.privateKeyToFile"

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(pk.Key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	filePath := filepath.Join(kDir, pk.ID+".pem")
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("%s: failed to create key file: %v", op, err)
	}
	defer file.Close()

	if err := pem.Encode(file, pemBlock); err != nil {
		return fmt.Errorf("%s: failed to encode private key: %v", op, err)
	}
	return nil
}

func deletePrivateKey(
	ks *KeysStore,
	pk *models.PrivateKey,
) error {
	const op = "jwks.keys.utils.deletePrivateKey"

	if err := deletePrivateKeyFile(pk.ID, ks.KeysDir); err != nil {
		if os.IsNotExist(err) {
			ks.Log.Debug("private key file not found, skipping deletion", slog.String("op", op), slog.String("keyID", pk.ID))
		} else {
			return fmt.Errorf("%s: %v", op, err)
		}
	}
	delete(ks.PrivateKeys, pk.ID)

	ks.Log.Debug("private key %s is deleted", pk.ID, slog.String("op", op))

	return nil
}

func deletePrivateKeyFile(keyID string, kDir string) error {
	const op = "jwks.keys.utils.deletePrivateKeyFile"

	fileName := fmt.Sprintf("%s.pem", keyID)
	path := path.Join(kDir, fileName)

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

// ----------------------------------End Block----------------------------------

// -------------------------------Public Key Block------------------------------

func getPublicKey(
	ks *KeysStore,
	privateKey *rsa.PrivateKey,
) error {
	panic("implement me!")
}

// Удаляет все все публичные ключи из keys folder
func deletePublicKeys(ks *KeysStore) error {
	const op = "utils.jwks.DeletePublicKeys"

	keys, err := os.ReadDir(ks.KeysDir)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	for _, key := range keys {
		if !key.IsDir() && path.Ext(key.Name()) == ".pub.pem" {
			os.Remove(key.Name())
		}
	}
	ks.Log.Debug("public keys successfully deleted", slog.String("op", op))
	return nil
}

// ----------------------------------End Block----------------------------------

// FIXME
// func saveKeys(
// 	ks *KeysStore,
// 	k *rsa.PrivateKey,
// 	id string,
// ) error {
// 	if ks.PrivateKey == nil {
// 		ks.PrivateKey = &models.PrivateKey{
// 			ID:  id,
// 			Key: k,
// 		}

// 		ks.PublicKeys[id] = &models.PublicKey{
// 			ID:  id,
// 			Key: &k.PublicKey,
// 		}

// 		return nil
// 	}

// 	deletePrivateKey()
// 	go deletePublicKey()

// }

func checkKeysFolder(ks *KeysStore) error {
	const op = "jwks.CheckKeysFolder"

	if _, err := os.Stat(ks.KeysDir); os.IsNotExist(err) {
		if err := createKeysFolder(ks); err != nil {
			return fmt.Errorf("%s: %v", op, err)
		}
	} else if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil

}

func createKeysFolder(ks *KeysStore) error {
	const op = "jwks.keys.utils.createKeysFolder"

	if err := os.Mkdir(ks.KeysDir, 0700); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	ks.Log.Debug("keys folder created", slog.String("op", op))
	return nil
}

// Перевод PublicKey в формат для передачи клиентам
func convertToJWKS(
	ks *KeysStore,
	pubKey *rsa.PublicKey,
) (map[string]any, error) {
	panic("implement me!")
}

// gorutine для отложенного удаления public key через определенное время
func deletePublicKeyTask(ks *KeysStore, id string) error {
	const op = "jwks.keys.utils.deletePublicKey"

	timer := time.NewTimer(ks.TokenTTL + time.Minute)

	go func() {
		<-timer.C
		delete(ks.PublicKeys, id)
	}()

	ks.Log.Debug("public key %s is deleted", id, slog.String("op", op))
	return nil
}

func listPrivateKeysFiles(keysDir string) ([]os.DirEntry, error) {
	const op = "jwks.keys.utils.getPrivateKeysFiles"

	entries, err := os.ReadDir(keysDir)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read directory: %v", op, err)
	}

	var pemFiles []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && path.Ext(entry.Name()) == ".pem" {
			pemFiles = append(pemFiles, entry)
		}
	}

	if len(pemFiles) == 0 {
		return nil, errKeyNotExist
	}

	return pemFiles, nil
}
