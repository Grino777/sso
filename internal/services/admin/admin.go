package admin

import "log/slog"

type AdminService struct {
	logger *slog.Logger
}

func NewAdminService(log *slog.Logger) *AdminService {
	return &AdminService{
		logger: log,
	}
}

func (as *AdminService) RotateJwksKeys() error {
	panic("implement me!")
}
