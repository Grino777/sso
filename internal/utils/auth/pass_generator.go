package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func CreatePassHash(pass string) (string, error) {
	const op = "utils.auth.createPassHash"

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("%s: %v", op, err)
	}
	return string(passHash), nil
}
