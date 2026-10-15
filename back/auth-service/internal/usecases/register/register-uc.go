package register

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
)

type RegisterUC struct {
	users        ports.IUserRepository
	hasher       ports.IPasswordHasher
	startSession *startsession.StartSessionUC
}

func NewRegisterUC(users ports.IUserRepository, hasher ports.IPasswordHasher, startSession *startsession.StartSessionUC) *RegisterUC {
	return &RegisterUC{users: users, hasher: hasher, startSession: startSession}
}

// Execute creates a user with an email and password and opens its session on deviceId.
func (r *RegisterUC) Execute(ctx context.Context, email, password, deviceId string) (*models.TokenPair, error) {
	if err := models.ValidateDeviceId(deviceId); err != nil {
		return nil, err
	}
	email, err := models.NormalizeEmail(email)
	if err != nil {
		return nil, err
	}
	if err := models.ValidatePassword(password); err != nil {
		return nil, err
	}

	hash, err := r.hasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := models.NewPasswordUser(email, hash)
	if err := r.users.Create(ctx, user); err != nil {
		return nil, err
	}
	return r.startSession.Execute(ctx, user, deviceId)
}
