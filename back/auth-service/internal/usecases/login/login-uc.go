package login

import (
	"context"
	"errors"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/startsession"
)

type LoginUC struct {
	users        ports.IUserRepository
	hasher       ports.IPasswordHasher
	startSession *startsession.StartSessionUC
}

func NewLoginUC(users ports.IUserRepository, hasher ports.IPasswordHasher, startSession *startsession.StartSessionUC) *LoginUC {
	return &LoginUC{users: users, hasher: hasher, startSession: startSession}
}

// Execute opens a session on deviceId for the user with email and password. Every mismatch
// returns models.ErrInvalidCredentials, so the response doesn't reveal which emails exist.
func (l *LoginUC) Execute(ctx context.Context, email, password, deviceId string) (*models.TokenPair, error) {
	if err := models.ValidateDeviceId(deviceId); err != nil {
		return nil, err
	}
	email, err := models.NormalizeEmail(email)
	if err != nil || models.ValidatePassword(password) != nil {
		return nil, models.ErrInvalidCredentials
	}

	user, err := l.users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, models.ErrUserNotFound) {
		return nil, err
	}
	if user == nil || !user.HasPassword() {
		// hashes anyway so an unknown email takes as long as a wrong password
		_, _ = l.hasher.Hash(password)
		return nil, models.ErrInvalidCredentials
	}

	ok, err := l.hasher.Verify(password, user.PasswordHash)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, models.ErrInvalidCredentials
	}
	return l.startSession.Execute(ctx, user, deviceId)
}
