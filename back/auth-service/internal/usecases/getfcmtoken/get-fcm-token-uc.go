package getfcmtoken

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

type GetFCMTokenUC struct {
	users ports.IUserRepository
}

func NewGetFCMTokenUC(users ports.IUserRepository) *GetFCMTokenUC {
	return &GetFCMTokenUC{users: users}
}

func (g *GetFCMTokenUC) Execute(ctx context.Context, userId int) (string, error) {
	user, err := g.users.GetById(ctx, userId)
	if err != nil {
		return "", err
	}
	if user.FCMToken == "" {
		return "", models.ErrFCMTokenNotFound
	}
	return user.FCMToken, nil
}
