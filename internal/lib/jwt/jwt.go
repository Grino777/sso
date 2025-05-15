package jwt

import (
	"time"

	"github.com/Grino777/sso/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
)

// Create new token for user
func NewAccessToken(user *models.User, app *models.App, d time.Duration) (*models.Token, error) {

	tObj := &models.Token{}

	token := jwt.New(jwt.SigningMethodRS256)
	expire_at := time.Now().UTC().Add(d).Unix()

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["role_id"] = user.Role_id
	claims["username"] = user.Username
	claims["app_id"] = app.ID
	claims["exp"] = expire_at

	tokenString, err := token.SignedString([]byte(app.Secret))
	if err != nil {
		return tObj, err
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
