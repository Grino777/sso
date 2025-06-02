package sqlite

import (
	"context"

	"github.com/Grino777/sso/internal/domain/models"
)

// FIXME
func (s *SQLiteStorage) SaveUserToken(ctx context.Context,
	user models.User,
	app models.App,
) error {
	panic("implement me!")
}
