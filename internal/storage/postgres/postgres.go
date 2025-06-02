package postgres

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Grino777/sso/internal/config"
	"github.com/Grino777/sso/internal/domain/models"
	"github.com/jackc/pgx/v5"
)

const pgOp = "storage.postgres.postgres."

type PostgresStorage struct {
	client *pgx.Conn
	logger *slog.Logger
	cfg    config.DatabaseConfig
}

func NewPostgresStorage(
	cfg config.DatabaseConfig,
	log *slog.Logger,
) *PostgresStorage {
	return &PostgresStorage{
		logger: log,
		cfg:    cfg,
	}
}

func (ps *PostgresStorage) Connect(ctx context.Context) error {

	user := ps.cfg.DBUser
	pass := ps.cfg.DBPass
	host := ps.cfg.DBHost
	port := ps.cfg.DBPort
	db := ps.cfg.DBName

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, pass, host, port, db)
	client, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to Postgres: %w", err)
	}

	ps.client = client

	return nil
}

func (ps *PostgresStorage) Close(ctx context.Context) error {
	const op = pgOp + "Close"

	if err := ps.client.Close(ctx); err != nil {
		return fmt.Errorf("%s: %v", op, err)
	}
	return nil
}

func (ps *PostgresStorage) SaveUser(ctx context.Context, user, passHash string) error {
	panic("implement me!")
}

func (ps *PostgresStorage) GetUser(ctx context.Context, username string) (models.User, error) {
	panic("implement me!")
}

func (ps *PostgresStorage) IsAdmin(ctx context.Context, username string) (bool, error) {
	panic("implement me!")
}

func (ps *PostgresStorage) GetApp(ctx context.Context, appID uint32) (models.App, error) {
	panic("implement me!")
}

func (ps *PostgresStorage) DeleteRefreshToken(ctx context.Context, userID uint64, appID uint32, token models.Token) error {
	panic("implement me!")
}

func (ps *PostgresStorage) SaveRefreshToken(ctx context.Context, userID uint64, appID uint32, token models.Token) error {
	panic("implement me!")
}
