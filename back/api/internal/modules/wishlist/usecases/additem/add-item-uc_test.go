package additem_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/additem"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
)

func TestAddItemUC_Execute_ValidItem_StoresAndNotifies(t *testing.T) {
	t.Parallel()

	// Arrange
	var stored *models.Item
	uc := additem.NewAddItemUC(&mocks.MockItemRepository{
		AddFn: func(ctx context.Context, item *models.Item) error {
			stored = item
			item.Id = 7
			return nil
		},
	})
	notified := 0
	uc.SetOnChange(func(userId int) { notified = userId })

	// Act
	item, err := uc.Execute(context.Background(), 3, models.ItemKindAuthor, "Carmen Mola", "", "", "OL9A")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(7), item.Id)
	assert.Equal(t, "OL9A", stored.AuthorKey)
	assert.Equal(t, 3, notified)
}

func TestAddItemUC_Execute_Failures_DoesNotNotify(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		count int
		add   error
		title string
		err   error
	}{
		{"Invalid", 0, nil, "", models.ErrEmptyTitle},
		{"Full", 500, nil, "x", models.ErrWishlistFull},
		{"Duplicate", 0, models.ErrDuplicate, "x", models.ErrDuplicate},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := additem.NewAddItemUC(&mocks.MockItemRepository{
				CountByUserIdFn: func(ctx context.Context, userId int) (int, error) { return tc.count, nil },
				AddFn:           func(ctx context.Context, item *models.Item) error { return tc.add },
			})
			notified := false
			uc.SetOnChange(func(int) { notified = true })

			// Act
			_, err := uc.Execute(context.Background(), 1, models.ItemKindBook, tc.title, "", "", "")

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.False(t, notified)
		})
	}
}
