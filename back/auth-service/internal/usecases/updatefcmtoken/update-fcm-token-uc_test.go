package updatefcmtoken_test

import (
	"context"
	"strings"
	"testing"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/updatefcmtoken"
	"github.com/stretchr/testify/assert"
)

func TestUpdateFCMTokenUC_Execute_Token_Result(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		token string
		saved string
		err   error
	}{
		{"Token", " fcm-1 ", "fcm-1", nil},
		{"Empty removes", "", "", nil},
		{"Too long", strings.Repeat("a", 4097), "unchanged", models.ErrInvalidFCMToken},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			saved := "unchanged"
			users := &mocks.MockUserRepository{
				UpdateFCMTokenFn: func(ctx context.Context, id int, token string) error {
					saved = token
					return nil
				},
			}

			// Act
			err := updatefcmtoken.NewUpdateFCMTokenUC(users).Execute(context.Background(), 1, tc.token)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.saved, saved)
		})
	}
}
