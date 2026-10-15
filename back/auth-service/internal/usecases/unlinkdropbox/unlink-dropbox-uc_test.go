package unlinkdropbox_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/unlinkdropbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnlinkDropboxUC_Execute_User_DeletesLink(t *testing.T) {
	t.Parallel()

	// Arrange
	var deleted []int
	links := &mocks.MockDropboxLinkRepository{
		DeleteFn: func(ctx context.Context, userId int) error {
			deleted = append(deleted, userId)
			return nil
		},
	}

	// Act
	err := unlinkdropbox.NewUnlinkDropboxUC(links).Execute(context.Background(), 4)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []int{4}, deleted)
}
