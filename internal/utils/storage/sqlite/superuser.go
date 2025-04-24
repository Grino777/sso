package sqlite

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

// Cоздает пользователя с ролью superadmin, если такой еще не существует.
func CreateSuperUser(db *sql.DB,
	username, password string,
) {
	query := "SELECT COUNT(*) FROM users WHERE role_id = 3"
	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		panic("error getting superuser from DB")
	}

	if count > 0 {
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic("error creating password for superuser")
	}

	query = "INSERT INTO users (username, pass_hash, role_id) VALUES (?, ?, 3)"
	_, err = db.Exec(query, username, string(hashedPassword))
	if err != nil {
		panic("error inserting superuser to DB")
	}
}
