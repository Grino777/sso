package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/Grino777/sso/internal/domain/models"
	"github.com/Grino777/sso/internal/lib/jwt"
	"github.com/Grino777/sso/internal/lib/logger"
	"github.com/Grino777/sso/internal/storage/sqlite"
	"golang.org/x/crypto/bcrypt"
)

const authUOp = "services.auth.utils."

func ValidateUser(username, password string) error {
	user := models.User{Username: username, Password: password}
	err := user.IsValid()
	return err
}

func ValidateApp(appID uint32) error {
	app := models.App{ID: appID}
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

func (s *AuthService) generateUserTokens(ctx context.Context, user models.User, app models.App) (models.User, error) {
	const op = authUOp + "generateUserTokens"

	log := s.Logger.With(slog.String("op", op), slog.String("username", user.Username))

	privateKey, err := s.KeysStore.GetLatestPrivateKey()
	if err != nil {
		return models.User{}, err
	}

	tokens, err := jwt.CreateNewTokens(user, app, privateKey, s.Tokens)
	if err != nil {
		log.Error("failed to create new tokens", logger.Error(err))
		return models.User{}, err
	}

	for attempts := 0; attempts < 10; attempts++ {
		if err := s.DB.SaveRefreshToken(ctx, user.ID, app.ID, tokens.RefreshToken); err != nil {
			if errors.Is(err, sqlite.ErrRefreshTokenExist) {
				log.Debug("refresh token already exists, generating new token")
				refreshToken, err := jwt.NewRefreshToken(s.Tokens.RefreshTokenTTL)
				if err != nil {
					log.Error("failed to generate new refresh token", logger.Error(err))
					return models.User{}, fmt.Errorf("%s: failed to generate new refresh token: %w", op, err)
				}
				tokens.RefreshToken = refreshToken
				continue // Повторяем попытку с новым токеном
			}
			log.Error("failed to save refresh token", logger.Error(err))
			return models.User{}, fmt.Errorf("%s: failed to save refresh token: %w", op, err)
		}
		log.Debug("refresh token updated")
		break
	}
	user.Tokens = tokens

	return user, nil
}

func (s *AuthService) validatePassword(passHash []byte, password string) error {
	const op = authUOp + "validatePassword"

	log := s.Logger.With(slog.String("op", op))

	if err := bcrypt.CompareHashAndPassword(passHash, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Error("invalid credentials", logger.Error(err))
			return ErrInvalidCredentials
		}
		log.Error("%s:%w", op, err)
		return err
	}
	return nil
}

func (s *AuthService) loggingUserLogin() {}
