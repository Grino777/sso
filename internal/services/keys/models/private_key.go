package models

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

var (
	ErrPrivateKeyExpired = errors.New("private key expired")
)

type PrivateKey struct {
	ID        string
	Key       *rsa.PrivateKey
	CreatedAt time.Time
	ExpireAt  time.Time
	publicKey *PublicKey
	filePath  string
	keyTTL    time.Duration
}

func NewPrivateKey(
	keyTTL time.Duration,
	tokenTTL time.Duration,
	keysDir string,
) (*PrivateKey, error) {
	const op = opKeys + "NewPrivateKey"

	rawPk, err := rsa.GenerateKey(rand.Reader, 3072)
	if err != nil {
		return nil, fmt.Errorf("%s: private key not generated: %w", op, err)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	s, ns := id.Time().UnixTime()
	createdAt := time.Unix(s, ns)
	expireAt := createdAt.Add(keyTTL)

	absFilePath := filepath.Join(keysDir, id.String()+".pem")

	pk := &PrivateKey{
		ID:        id.String(),
		Key:       rawPk,
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
		filePath:  absFilePath,
		keyTTL:    keyTTL,
	}

	publicKey, err := pk.ConvertToPublic(tokenTTL, keysDir)
	if err != nil {
		return nil, err
	}

	pk.publicKey = publicKey
	return pk, nil
}

func (pk *PrivateKey) GetPublicKey() *PublicKey {
	return pk.publicKey
}

func (pk *PrivateKey) SaveToFile(keysDir string) error {
	const op = opKeys + "SaveToFile"

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(pk.Key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	filePath := filepath.Join(keysDir, pk.ID+".pem")
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

func (pk *PrivateKey) CheckExpiration() bool {
	now := time.Now()
	endedTime := pk.CreatedAt.Add(pk.keyTTL)

	if endedTime.After(now) {
		return false
	}
	return true
}

func (pk *PrivateKey) ConvertToPublic(tokenTTL time.Duration, keysDir string) (*PublicKey, error) {
	publicKey, err := NewPublicKey(pk, tokenTTL, keysDir)
	if err != nil {
		return nil, err
	}
	return publicKey, nil
}

func (pk *PrivateKey) DeletePemFile() error {
	if err := os.Remove(pk.filePath); err != nil {
		return fmt.Errorf("failed to remove pem file for key: %s", pk.ID)
	}
	return nil
}

func DecodePemToPrivateKey(
	keyFileName, keysDir string,
	keyTTL, tokenTTL time.Duration,
) (*PrivateKey, error) {
	const op = opKeys + "Decode"

	absKeyPath := filepath.Join(keysDir, keyFileName)
	privatePEM, err := os.ReadFile(absKeyPath)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to read private key %s: %w", op, keyFileName, err)
	}

	privateBlock, _ := pem.Decode(privatePEM)
	if privateBlock == nil {
		return nil, fmt.Errorf("%s: failed to decode private key PEM", op)
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(privateBlock.Bytes)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to parse private key: %s, %w", op, keyFileName, err)
	}

	keyID := strings.TrimSuffix(keyFileName, ".pem")
	keyUUID, err := uuid.Parse(keyID)
	if err != nil {
		return nil, err
	}

	sec, ns := keyUUID.Time().UnixTime()
	createdAt := time.Unix(sec, ns)
	expireAt := createdAt.Add(keyTTL)

	keyInstance := &PrivateKey{
		ID:        keyID,
		Key:       privateKey,
		CreatedAt: createdAt,
		ExpireAt:  expireAt,
		filePath:  absKeyPath,
		keyTTL:    keyTTL,
	}
	publicKey, err := NewPublicKey(keyInstance, tokenTTL, keysDir)
	if err != nil {
		return keyInstance, err
	}

	keyInstance.publicKey = publicKey
	return keyInstance, nil
}
