package sqlite

import (
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Cоздает пользователя с ролью superadmin, если такой еще не существует.
func CreateSuperUser(db *sql.DB,
	username, password string,
) error {
	const op = "storage.sqlite.CreateSuperUser"

	query := "SELECT COUNT(*) FROM users WHERE role_id = 3"
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return fmt.Errorf("error getting superuser from DB: %v", err)
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: error creating password for superuser: %v", op, err)
	}

	query = "INSERT INTO users (username, pass_hash, role_id) VALUES (?, ?, 3)"
	_, err = db.Exec(query, username, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("%s: error inserting superuser to DB: %v", op, err)
	}

	return nil
}
