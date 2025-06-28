package store

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/lib/logger"
)

const (
	tokenTTL = 5 * time.Second
	keyTTL   = 5 * time.Second
)

func TestManager__NewStore(t *testing.T) {
	log := logger.NewLogger(os.Stdout, slog.LevelDebug)
	cfgPath := config.PathConfig{KeysDir: "/home/grino777/project/go/sso/test_keys"}
	cfgTTL := config.TTLConfig{
		TokenTTL: tokenTTL,
		KeyTTL:   keyTTL,
	}

	ks, err := NewKeysStore(log, cfgPath, cfgTTL)
	t.Log("new keys store created")
	if err != nil {
		t.Fatal("failed to create new keys store")
	}

	// Test RotateKeys function
	t.Run("TestRotateKeys", func(t *testing.T) {
		log.Info("start TestRotateKeys")

		oldKey, err := ks.GetLatestPrivateKey()
		if err != nil {
			t.Error("failed to get latest private key")
		}
		t.Log("old key:", oldKey)

		_, exists := ks.PublicKeys[oldKey.ID]
		if !exists {
			t.Errorf("old key does not exist before rotating keys: %s", oldKey.ID)
		}

		// Rotate keys and check if a new key pair is generated
		keys, err := ks.RotateKeys()
		if err != nil {
			t.Errorf("failed to rotate keys: %v", err)
		}
		if oldKey.ID == keys.PrivateKey.ID {
			t.Errorf("new key matches the old one during rotation")
		}

		newKey, err := ks.GetLatestPrivateKey()
		if err != nil {
			t.Error("failed to get latest private key after rotation")
		}
		if oldKey.ID == newKey.ID {
			t.Error("new key matches the old one during rotation")
		}
		_, exist := ks.PublicKeys[oldKey.ID]
		if !exist {
			t.Errorf("old key is missing in the public keys map after rotation: %s", oldKey.ID)
		}
		_, exist = ks.PublicKeys[newKey.ID]
		if !exist {
			t.Errorf("new key is missing in the public keys map after rotation: %s", newKey.ID)
		}
	})

	// Test expiration of private key TTL
	t.Run("TestExperationPrivateKey", func(t *testing.T) {
		oldKey, err := ks.GetLatestPrivateKey()
		if err != nil {
			t.Errorf("error: %v", err)
		}

		time.Sleep(tokenTTL)
		newKey, err := ks.GetLatestPrivateKey()
		if err != nil {
			t.Errorf("failed to get public keys after token TTL expired: %v", err)
		}
		if oldKey.ID == newKey.ID {
			t.Error("new key matches the old one during rotation")
		}
	})

}
