package admin

import (
	"log/slog"
)

const adminOp = "services.admin."

type AdminService struct {
	logger *slog.Logger
}

func NewAdminService(log *slog.Logger) *AdminService {
	return &AdminService{
		logger: log,
	}
}
