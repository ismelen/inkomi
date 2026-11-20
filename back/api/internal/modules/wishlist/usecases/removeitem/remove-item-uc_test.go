package removeitem_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/removeitem"
)

func TestRemoveItemUC_Execute_Results_DeletesTheItemOfTheUser(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
	}{
		{"Removed", nil},
		{"Not found", models.ErrItemNotFound},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var gotId int64
			var gotUser int
			uc := removeitem.NewRemoveItemUC(&mocks.MockItemRepository{
				DeleteFn: func(ctx context.Context, id int64, userId int) error {
					gotId, gotUser = id, userId
					return tc.err
				},
			})

			// Act
			err := uc.Execute(context.Background(), 1, 9)

			// Assert
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, int64(9), gotId)
			assert.Equal(t, 1, gotUser)
		})
	}
}
