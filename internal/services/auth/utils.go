package auth

import (
	"context"

	"github.com/Grino777/sso/internal/domain/models"
)

func ValidateUser(username, password string) error {
	user := models.User{Username: username, Password: password}
	err := user.IsValid()
	return err
}

func ValidateApp(appID uint32) error {
	app := models.App{ID: int(appID)}
	err := app.IsValid()
	return err
}

func ValidateData(
	ctx context.Context,
	username, password string,
	appID uint32,
) error {
	err := ValidateUser(username, password)
	if err != nil {
		return err
	}

	err = ValidateApp(appID)
	if err != nil {
		return err
	}

	return nil
}
