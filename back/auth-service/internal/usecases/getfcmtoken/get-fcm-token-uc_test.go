package getfcmtoken_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/getfcmtoken"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetFCMTokenUC_Execute_User_Result(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		users *mocks.MockUserRepository
		token string
		err   error
	}{
		{"Token", &mocks.MockUserRepository{
			GetByIdFn: func(ctx context.Context, id int) (*models.User, error) { return &models.User{FCMToken: "fcm"}, nil },
		}, "fcm", nil},
		{"No token", &mocks.MockUserRepository{
			GetByIdFn: func(ctx context.Context, id int) (*models.User, error) { return &models.User{}, nil },
		}, "", models.ErrFCMTokenNotFound},
		{"Unknown user", &mocks.MockUserRepository{}, "", models.ErrUserNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			token, err := getfcmtoken.NewGetFCMTokenUC(tc.users).Execute(context.Background(), 1)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.token, token)
		})
	}
}
