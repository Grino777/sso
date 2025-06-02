package models

import (
	"crypto/rsa"

	"github.com/Grino777/sso-proto/gen/go/sso"
)

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

type JWKSToken struct {
	Kid string
	Kty string
	Alg string
	Use string
	N   string
	E   string
}

func (jt *JWKSToken) ConvertToken() *sso.Jwk {
	return &sso.Jwk{
		Kid: jt.Kid,
		Kty: jt.Kty,
		Use: jt.Use,
		N:   jt.N,
		E:   jt.E,
	}
}
