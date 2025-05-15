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

	"github.com/Grino777/sso/internal/services/jwks/models"
)

var (
	ErrNotExist = errors.New("private key not exist")
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

	if err := getPrivateKey(ks); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	if err := getPublicKey(ks); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

// ----------------------------------End Block----------------------------------

// ------------------------------Private Key Block------------------------------

func getPrivateKey(ks *KeysStore) error {
	const op = "jwks.keys.getPrivateKey"

	privateKey, err := privateKeyFile(ks)
	if err != nil {
		if errors.Is(err, ErrNotExist) {
			ks.log.Debug("private key not found")
			if err := genPrivateKey(ks); err != nil {
				return fmt.Errorf("%s: %v", op, err)
			}
			return nil
		}
		return fmt.Errorf("%s: %v", op, err)
	}
	ks.PrivateKey = privateKey

	if err := deletePublicKeys(ks); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	ks.log.Debug("private key successfully received", slog.String("op", op))
	return nil
}

func genPrivateKey(ks *KeysStore) error {
	const op = "jwks.keys.genPrivateKey"

	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("%s: private key not generated: %v", op, err)
	}

	if err := saveKeys(ks, pk); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

func privateKeyFile(ks *KeysStore) (*models.PrivateKey, error) {
	const op = "utils.jwks.GetPrivateKey"

	entries, err := os.ReadDir(ks.keysDir)
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
		return nil, ErrNotExist
	}

	sort.Slice(pemFiles, func(i, j int) bool {
		infoI, _ := pemFiles[i].Info()
		infoJ, _ := pemFiles[j].Info()
		return infoI.ModTime().After(infoJ.ModTime())
	})

	latestKeyPath := path.Join(ks.keysDir, pemFiles[0].Name())
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
		oldKeyPath := path.Join(ks.keysDir, entry.Name())
		if err := os.Remove(oldKeyPath); err != nil {
			return nil, fmt.Errorf("%s: failed to remove old key %s: %v", op, oldKeyPath, err)
		}
	}

	return &models.PrivateKey{
		ID:  keyID,
		Key: privateKey,
	}, nil
}

func setPrivateKey(ks *KeysStore) error {
	panic("implement me!")
}

// ----------------------------------End Block----------------------------------

// -------------------------------Public Key Block------------------------------

func getPublicKey(ks *KeysStore) error {
	panic("implement me!")
}

func deletePublicKeys(ks *KeysStore) error {
	const op = "utils.jwks.DeletePublicKeys"

	keys, err := os.ReadDir(ks.keysDir)
	if err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	for _, key := range keys {
		if !key.IsDir() && path.Ext(key.Name()) == ".pub.pem" {
			os.Remove(key.Name())
		}
	}
	ks.log.Debug("public keys successfully deleted", slog.String("op", op))
	return nil
}

// ----------------------------------End Block----------------------------------

// FIXME
func saveKeys(
	ks *KeysStore,
	k *rsa.PrivateKey,
) error {
	// publicKey := k.PublicKey
	panic("implement me!")

}

func checkKeysFolder(ks *KeysStore) error {
	const op = "jwks.CheckKeysFolder"

	if _, err := os.Stat(ks.keysDir); os.IsNotExist(err) {
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

	if err := os.Mkdir(ks.keysDir, 0700); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	ks.log.Debug("keys folder created", slog.String("op", op))

	if err := ks.RotateKeys(); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}

	return nil
}

func toJWKS(
	ks *KeysStore,
	pubKey *rsa.PublicKey,
) (map[string]any, error) {
	panic("implement me!")
}
