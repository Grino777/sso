package auth

import sso_v1 "github.com/Grino777/sso-proto/gen/go/sso"

type serverAPI struct {
	sso_v1.UnimplementedAuthServer
	auth Auth
}

type Auth interface {
}
