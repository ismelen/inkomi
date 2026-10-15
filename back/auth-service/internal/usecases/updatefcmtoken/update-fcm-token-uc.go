package updatefcmtoken

import (
	"context"
	"strings"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
)

const maxFCMTokenLength = 4096

type UpdateFCMTokenUC struct {
	users ports.IUserRepository
}

func NewUpdateFCMTokenUC(users ports.IUserRepository) *UpdateFCMTokenUC {
	return &UpdateFCMTokenUC{users: users}
}

// Execute replaces the FCM token of the user; an empty token removes it.
func (u *UpdateFCMTokenUC) Execute(ctx context.Context, userId int, token string) error {
	token = strings.TrimSpace(token)
	if len(token) > maxFCMTokenLength {
		return models.ErrInvalidFCMToken
	}
	return u.users.UpdateFCMToken(ctx, userId, token)
}
