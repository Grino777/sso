package models

import (
	"github.com/Grino777/sso-proto/gen/go/sso"
)

const opKeys = "keys.models."

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
