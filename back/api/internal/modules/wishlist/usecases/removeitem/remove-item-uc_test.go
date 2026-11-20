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

func TestRemoveItemUC_Execute_Results_NotifiesOnlyOnSuccess(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		err      error
		notified bool
	}{
		{"Removed", nil, true},
		{"Not found", models.ErrItemNotFound, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var gotId int64
			uc := removeitem.NewRemoveItemUC(&mocks.MockItemRepository{
				DeleteFn: func(ctx context.Context, id int64, userId int) error {
					gotId = id
					return tc.err
				},
			})
			notified := false
			uc.SetOnChange(func(int) { notified = true })

			// Act
			err := uc.Execute(context.Background(), 1, 9)

			// Assert
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, int64(9), gotId)
			assert.Equal(t, tc.notified, notified)
		})
	}
}
