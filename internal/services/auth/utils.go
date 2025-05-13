package auth

import (
	"context"

	"github.com/Grino777/sso/internal/domain/models"
)

func ValidateUser(username, password string) (user *models.User, err error) {
	user = &models.User{Username: username, Password: password}
	err = user.IsValid()
	return user, err
}

func ValidateApp(appID uint32) (app *models.App, err error) {
	app = &models.App{ID: int(appID)}
	err = app.IsValid()
	return app, err
}

func ValidateData(
	ctx context.Context,
	username, password string,
	appID uint32,
) (*models.User, *models.App, error) {
	user, err := ValidateUser(username, password)
	if err != nil {
		return nil, nil, err
	}

	app, err := ValidateApp(appID)
	if err != nil {
		return nil, nil, err
	}

	return user, app, nil
}
