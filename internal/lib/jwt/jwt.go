package jwt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models"
	jwksM "github.com/Grino777/sso/internal/services/jwks/models"

	"github.com/golang-jwt/jwt/v5"
)

func CreateNewTokens(
	user models.User,
	app models.App,
	pk *jwksM.PrivateKey,
	tokens config.TokenConfig,
) (models.Tokens, error) {
	acessToken, err := NewAccessToken(user, app, pk, tokens.TokenTTL)
	if err != nil {
		return models.Tokens{}, err
	}

	refreshToken, err := NewRefreshToken(tokens.RefreshTokenTTL)
	if err != nil {
		return models.Tokens{}, err
	}

	return models.Tokens{
		AccessToken:  acessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Create new token for user
func NewAccessToken(
	user models.User,
	app models.App,
	pk *jwksM.PrivateKey,
	d time.Duration,
) (models.Token, error) {
	const op = "lib.jwt.NewAccessToken"

	tObj := models.Token{}

	token := jwt.New(jwt.SigningMethodRS256)
	expire_at := time.Now().UTC().Add(d).Unix()

	claims := token.Claims.(jwt.MapClaims)
	claims["kid"] = pk.ID
	claims["user_id"] = user.ID
	claims["role_id"] = user.Role_id
	claims["username"] = user.Username
	claims["app_id"] = app.ID
	claims["exp"] = expire_at

	tokenString, err := token.SignedString(pk.Key)
	if err != nil {
		return tObj, fmt.Errorf("%s: %w", op, err)
	}

	tObj = models.Token{
		Token:     tokenString,
		Expire_at: expire_at,
	}

	return tObj, nil
}

func NewRefreshToken(d time.Duration) (models.Token, error) {
	const op = "lib.jwt.NewRefreshToken"
	const tokenLenght = 32

	tokenBytes := make([]byte, tokenLenght)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return models.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	tokenString := base64.URLEncoding.EncodeToString(tokenBytes)
	expire_at := time.Now().UTC().Add(d).Unix()

	token := models.Token{
		Token:     tokenString,
		Expire_at: expire_at,
	}
	return token, nil
}
