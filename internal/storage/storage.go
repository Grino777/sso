package storage

import (
	"context"
	"database/sql"
	"sso/migrations"
)

type Storage struct {
	driverName string
	db         *sql.DB
}

// Creates a new DB session and performs migrations
func New(driverName string, dbPath string) *Storage {
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	s := Storage{
		driverName: driverName,
		db:         db,
	}

	migrations.Migrate(s.db, driverName)

	return &s
}

// Close DB session
func (s *Storage) Close() error {
	return s.db.Close()
}

func (s *Storage) SaveUser(
	ctx context.Context,
	username string,
	password string,
) (token string, err error) {
	panic("implemented me")
}

func (s *Storage) GetUser(
	ctx context.Context,
	username string,
) (id int, err error) {
	panic("implement me")
}

func (s *Storage) GetApp(
	ctx context.Context,
	appID int,
) (appId int, err error) {
	panic("implemente me")
}

func (s *Storage) IsAdmin(
	ctx context.Context,
	username string,
) (isAdmin bool, err error) {
	panic("implement me")
}
