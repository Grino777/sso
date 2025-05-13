package models

import "fmt"

const (
	EmptyField         = "filed cannot be empty"
	FieldContainSpaces = "field cannot contain spaces"
	ShortPassword      = "password must be at least 6 characters"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error for field %q: %s", e.Field, e.Message)
}
