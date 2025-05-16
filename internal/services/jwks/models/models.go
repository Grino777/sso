package models

import "crypto/rsa"

type PrivateKey struct {
	ID      string
	Key     *rsa.PrivateKey
	IsSaved bool
}

type PublicKey struct {
	ID      string
	Key     *rsa.PublicKey
	IsSaved bool
}
