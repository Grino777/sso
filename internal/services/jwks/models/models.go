package models

import "crypto/rsa"

type PrivateKey struct {
	ID  string
	Key *rsa.PrivateKey
}

type PublicKey struct {
	ID  string
	Key *rsa.PublicKey
}
