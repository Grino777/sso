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

// Производит замену ключей (приватного и публичного)
// func (as *AdminService) RotateKeys() error {
// 	const op = adminOp + "RotateKeys"

// 	ctx := context.Background()

// 	oldPrivateKey, err := as.keyStore.GetLatestPrivateKey(ks)
// 	if err != nil {
// 		return err
// 	}
// 	if err := deletePrivateKey(ks, oldPrivateKey); err != nil {
// 		return err
// 	}
// 	newPrivateKey, err := generatePrivateKey(ks)
// 	if err != nil {
// 		return err
// 	}
// 	if err := setPrivateKey(ks, newPrivateKey); err != nil {
// 		return err
// 	}
// 	if err := deletePublicKeyTask(ctx, ks, oldPrivateKey.ID); err != nil {
// 		return err
// 	}

// 	ks.Log.Info("keys rotated successfully", slog.String("op", op), slog.String("newKeyID", newPrivateKey.ID))
// 	return nil
// }
