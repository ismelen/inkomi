package logout_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/logout"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogoutUC_Execute_Token_DeletesSessionByHash(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		token   string
		deleted []string
	}{
		{"Token", "refresh-1", []string{"hash:refresh-1"}},
		{"Empty token", "", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var deleted []string
			sessions := &mocks.MockSessionRepository{
				DeleteByRefreshHashFn: func(ctx context.Context, hash string) error {
					deleted = append(deleted, hash)
					return nil
				},
			}

			// Act
			err := logout.NewLogoutUC(sessions, &mocks.MockTokenIssuer{}).Execute(context.Background(), tc.token)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.deleted, deleted)
		})
	}
}
