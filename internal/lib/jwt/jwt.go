package jwt

import (
	"fmt"
	"time"

	"github.com/Grino777/sso/internal/domain/models"
	jwksM "github.com/Grino777/sso/internal/services/jwks/models"

	"github.com/golang-jwt/jwt/v5"
)

// Create new token for user
func NewAccessToken(
	user *models.User,
	app *models.App,
	pk *jwksM.PrivateKey,
	d time.Duration,
) (*models.Token, error) {
	const op = "lib.jwt.NewAccessToken"

	tObj := &models.Token{}

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
		return tObj, fmt.Errorf("%s: %v", op, err)
	}

	tObj = &models.Token{
		Token:     tokenString,
		Expire_at: expire_at,
	}

	return tObj, nil
}

// func NewRefreshToken(
// 	user models.User,
// 	app models.App,
// 	d time.Duration,
// ) (models.Token, error) {
// 	tokenObj := &models.Token{}
// 	expire_at := time.Now().UTC().Add(d).Unix()

// 	token := jwt.New(jwt.SigningMethodHS256)
// 	claims := token.Claims.(jwt.MapClaims)
// 	claims["user_id"] = user.ID
// 	claims["app_id"] = app.ID
// 	claims["exp"] = expire_at

// 	ts, err := token.SignedString()
// }
