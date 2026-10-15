package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
)

type IUserRepository interface {
	// Create stores user and sets its Id; it returns models.ErrEmailTaken if the email exists.
	Create(ctx context.Context, user *models.User) error
	GetById(ctx context.Context, id int) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	// Update saves the email, verification and password hash of user.
	Update(ctx context.Context, user *models.User) error
	UpdateFCMToken(ctx context.Context, id int, token string) error
}
