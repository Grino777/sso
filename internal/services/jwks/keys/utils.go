package keys

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Grino777/sso/internal/services/jwks/models"
	"github.com/google/uuid"
)

const utilsOp = "jwks.keys.utils."

var (
	errKeyNotExist = errors.New("private key not exist")
)

// ----------------------------------Init Block---------------------------------

func initKeys(ks *KeysStore) error {
	const op = utilsOp + "initKeys"

	if err := loadKeys(ks); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func loadKeys(ks *KeysStore) error {
	const op = utilsOp + "loadKeys"

	if err := checkKeysFolder(ks); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	privateKey, err := retrievePrivateKey(ks)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if err := setPrivateKey(ks, privateKey); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// ----------------------------------End Block----------------------------------

// ------------------------------Private Key Block------------------------------

func retrievePrivateKey(ks *KeysStore) (*models.PrivateKey, error) {
	const op = utilsOp + "retrievePrivateKey"

	privateKey, err := getLatestPrivateKey(ks)
	if err != nil {
		return nil, err
	}
	if err := deletePublicKeysFiles(ks); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	ks.Log.Debug("private key successfully received", slog.String("op", op))
	return privateKey, nil
}

func getLatestPrivateKey(ks *KeysStore) (*models.PrivateKey, error) {
	const op = utilsOp + "getLatestPrivateKey"

	pemFiles, err := listPrivateKeysFiles(ks.KeysDir)
	if errors.Is(err, errKeyNotExist) {
		privateKey, err := generatePrivateKey(ks)
		if err != nil {
			return nil, err
		}
		return privateKey, nil
	} else if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	latestKey, err := findLatestPrivateKey(ks, pemFiles)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := deleteOldPrivateKeys(ks, pemFiles); err != nil {
		return nil, err
	}

	return latestKey, nil
}

func deleteOldPrivateKeys(ks *KeysStore, pemFiles []os.DirEntry) error {
	const op = utilsOp + "deleteOldPrivateKeys"

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

func findLatestPrivateKey(ks *KeysStore, pemFiles []os.DirEntry) (*models.PrivateKey, error) {
	const op = utilsOp + "findLatestPrivateKey"

	sort.Slice(pemFiles, func(i, j int) bool {
		infoI, _ := pemFiles[i].Info()
		infoJ, _ := pemFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	latestKeyPath := filepath.Join(ks.KeysDir, pemFiles[0].Name())
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

func generatePrivateKey(ks *KeysStore) (*models.PrivateKey, error) {
	const op = utilsOp + "generatePrivateKey"

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

	if err := savePrivateKeyToFile(ks, pkObj); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return pkObj, nil
}

func setPrivateKey(ks *KeysStore, privateKey *models.PrivateKey) error {
	const op = utilsOp + "setPrivateKey"

	ks.mu.Lock()
	defer ks.mu.Unlock()

	if len(ks.PrivateKeys) > 1 {
		for _, v := range ks.PrivateKeys {
			if err := deletePrivateKey(ks, v); err != nil {
				return fmt.Errorf("%s: %w", op, err)
			}
		}
	}

	ks.PrivateKeys[privateKey.ID] = privateKey

	if !privateKey.IsSaved {
		if err := savePrivateKeyToFile(ks, privateKey); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
		privateKey.IsSaved = true
	}
	if err := savePublicKeyToFile(ks, privateKey); err != nil {
		return fmt.Errorf("%s: failed to save public key: %w", op, err)
	}

	return nil
}

func savePrivateKeyToFile(ks *KeysStore, privateKey *models.PrivateKey) error {
	const op = utilsOp + "savePrivateKeyToFile"

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey.Key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	filePath := filepath.Join(ks.KeysDir, privateKey.ID+".pem")
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

func deletePrivateKey(
	ks *KeysStore,
	privateKey *models.PrivateKey,
) error {
	const op = utilsOp + "deletePrivateKey"

	if err := deletePrivateKeyFile(privateKey.ID, ks.KeysDir); err != nil {
		if os.IsNotExist(err) {
			ks.Log.Debug("private key file not found, skipping deletion", slog.String("op", op), slog.String("keyID", privateKey.ID))
		} else {
			return fmt.Errorf("%s: %w", op, err)
		}
	}

	ks.mu.Lock()
	defer ks.mu.Unlock()
	delete(ks.PrivateKeys, privateKey.ID)

	ks.Log.Debug("private key %s is deleted", privateKey.ID, slog.String("op", op))

	return nil
}

func deletePrivateKeyFile(keyID string, kDir string) error {
	const op = utilsOp + "deletePrivateKeyFile"

	fileName := fmt.Sprintf("%s.pem", keyID)
	path := path.Join(kDir, fileName)

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
func deletePublicKeysFiles(ks *KeysStore) error {
	const op = utilsOp + "deletePublicKeys"

	entries, err := os.ReadDir(ks.KeysDir)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	var errs []error
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".pub.pem") {
			if err := os.Remove(filepath.Join(ks.KeysDir, entry.Name())); err != nil {
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
	ks.Log.Info("public keys deleted", slog.String("op", op))
	return nil
}

func savePublicKeyToFile(ks *KeysStore, privateKey *models.PrivateKey) error {
	const op = utilsOp + "savePublicKeyToFile"

	filePath := filepath.Join(ks.KeysDir, privateKey.ID+".pub.pem")

	pubASN1 := x509.MarshalPKCS1PublicKey(&privateKey.Key.PublicKey)
	pubPEM := &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: pubASN1,
	}
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("%s: failed to create key file: %w", op, err)
	}
	defer file.Close()

	if err := pem.Encode(file, pubPEM); err != nil {
		return fmt.Errorf("%s: failed to write public key to file: %w", op, err)
	}

	ks.PublicKeys[privateKey.ID] = &models.PublicKey{
		ID:      privateKey.ID,
		Key:     &privateKey.Key.PublicKey,
		IsSaved: true,
	}
	return nil
}

// ----------------------------------End Block----------------------------------

func checkKeysFolder(ks *KeysStore) error {
	const op = utilsOp + "checkKeysFolder"

	if _, err := os.Stat(ks.KeysDir); os.IsNotExist(err) {
		if err := createKeysFolder(ks); err != nil {
			return fmt.Errorf("%s: %w", op, err)
		}
	} else if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil

}

func createKeysFolder(ks *KeysStore) error {
	const op = utilsOp + "createKeysFolder"

	if err := os.Mkdir(ks.KeysDir, 0700); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	ks.Log.Debug("keys folder created", slog.String("op", op))
	return nil
}

// Перевод PublicKey в формат для передачи клиентам
func convertPubKeyToJWKS(pubKeyObj *models.PublicKey) (*models.JWKSToken, error) {
	pubKey := pubKeyObj.Key
	n := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes())

	return &models.JWKSToken{
		Kid: pubKeyObj.ID,
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		N:   n,
		E:   e,
	}, nil
}

// gorutine для отложенного удаления public key через определенное время
func deletePublicKeyTask(ctx context.Context, ks *KeysStore, id string) error {
	const op = utilsOp + "deletePublicKeyTask"

	timer := time.NewTimer(ks.TokenTTL + time.Minute)
	defer timer.Stop()

	go func() {
		select {
		case <-timer.C:
			ks.mu.Lock()
			defer ks.mu.Unlock()
			if _, exists := ks.PublicKeys[id]; exists {
				delete(ks.PublicKeys, id)
				ks.Log.Info("public key deleted", slog.String("op", op), slog.String("keyID", id))
			} else {
				ks.Log.Debug("public key not found, skipping deletion", slog.String("op", op), slog.String("keyID", id))
			}
		case <-ctx.Done():
			ks.Log.Debug("public key deletion cancelled", slog.String("op", op), slog.String("keyID", id))
			return
		}
	}()

	return nil
}

func listPrivateKeysFiles(keysDir string) ([]os.DirEntry, error) {
	const op = utilsOp + "listPrivateKeysFiles"

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
