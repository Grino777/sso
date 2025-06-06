package models

import (
	"strings"
)

const (
	MinLenPass = 5
)

type User struct {
	ID       uint64
	Username string
	Password string `json:"-"`
	PassHash []byte
	Role_id  int
	Tokens   Tokens
}

func (u *User) validateFields() error {
	if u.Username == "" {
		return &ValidationError{Field: "username", Message: EmptyField}
	}

	if strings.Contains(u.Username, " ") {
		return &ValidationError{Field: "username", Message: FieldContainSpaces}
	}

	if u.Password == "" {
		return &ValidationError{Field: "password", Message: EmptyField}
	}

	if len(u.Password) < MinLenPass {
		return &ValidationError{Field: "password", Message: ShortPassword}
	}

	return nil
}

func (u *User) IsValid() error {
	return u.validateFields()
}
