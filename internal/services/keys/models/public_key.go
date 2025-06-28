package models

import (
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"math/big"
	"time"
)

var (
	ErrPublicKeyExpired = errors.New("public key expired")
)

type PublicKey struct {
	ID          string
	Key         *rsa.PublicKey
	CreatedAt   time.Time
	ExpireAt    time.Time
	prirvateKey *PrivateKey
}

func NewPublicKey(privateKey *PrivateKey, tokenTTL time.Duration, keysDir string) (*PublicKey, error) {
	publicKey := &PublicKey{
		ID:          privateKey.ID,
		Key:         &privateKey.Key.PublicKey,
		CreatedAt:   privateKey.CreatedAt,
		ExpireAt:    privateKey.ExpireAt.Add(tokenTTL),
		prirvateKey: privateKey,
	}

	if publicKey.IsExpired() {
		return nil, ErrPublicKeyExpired
	}
	return publicKey, nil
}

func (pk *PublicKey) IsExpired() bool {
	now := time.Now().Local()

	if pk.ExpireAt.After(now) {
		return false
	}
	return true
}

func (pk *PublicKey) DeletePair() error {
	if err := pk.prirvateKey.DeletePemFile(); err != nil {
		return err
	}
	return nil
}

func (pk *PublicKey) ConvertToJWKS() *JWKSToken {
	pubKey := pk.Key
	n := base64.RawURLEncoding.EncodeToString(pubKey.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pubKey.E)).Bytes())

	return &JWKSToken{
		Kid: pk.ID,
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		N:   n,
		E:   e,
	}
}
