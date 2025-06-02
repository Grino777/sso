package models

type App struct {
	ID     uint32
	Name   string
	Secret string
}

func (a *App) validateFields() error {
	if a.ID == 0 {
		return &ValidationError{Field: "id", Message: EmptyField}
	}

	return nil
}

func (a *App) IsValid() error {
	return a.validateFields()
}
