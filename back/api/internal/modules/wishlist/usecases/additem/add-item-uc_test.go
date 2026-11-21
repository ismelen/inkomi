package additem_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/additem"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/resolvework"
)

func newUC(items *mocks.MockItemRepository, works *mocks.MockWorkRepository, catalog *mocks.MockCatalogLookup) *additem.AddItemUC {
	return additem.NewAddItemUC(items, resolvework.NewResolveWorkUC(items, works, catalog))
}

func TestAddItemUC_Execute_ValidItem_StoresIt(t *testing.T) {
	t.Parallel()

	// Arrange
	var stored *models.Item
	uc := newUC(&mocks.MockItemRepository{
		AddFn: func(ctx context.Context, item *models.Item) error {
			stored = item
			item.Id = 7
			return nil
		},
	}, &mocks.MockWorkRepository{}, &mocks.MockCatalogLookup{})

	// Act
	item, err := uc.Execute(context.Background(), 3, models.ItemKindAuthor, "Carmen Mola", "", "", "OL9A", "")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, int64(7), item.Id)
	assert.Equal(t, "OL9A", stored.AuthorKey)
}

func TestAddItemUC_Execute_Failures_ReturnsTheError(t *testing.T) {
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
			uc := newUC(&mocks.MockItemRepository{
				CountByUserIdFn: func(ctx context.Context, userId int) (int, error) { return tc.count, nil },
				AddFn:           func(ctx context.Context, item *models.Item) error { return tc.add },
			}, &mocks.MockWorkRepository{}, &mocks.MockCatalogLookup{})

			// Act
			_, err := uc.Execute(context.Background(), 1, models.ItemKindBook, tc.title, "", "", "", "")

			// Assert
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestAddItemUC_Execute_BookWithCover_ResolvesItsWork(t *testing.T) {
	t.Parallel()

	// Arrange
	var saved models.WorkDetails
	cover := "https://covers.openlibrary.org/b/id/7-M.jpg"
	uc := newUC(&mocks.MockItemRepository{}, &mocks.MockWorkRepository{
		SaveFn: func(ctx context.Context, work models.WorkDetails) error {
			saved = work
			return nil
		},
	}, &mocks.MockCatalogLookup{
		WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
			return models.WorkDetails{Title: "The Final Empire", Description: "Ash falls."}, nil
		},
	})

	// Act
	item, err := uc.Execute(context.Background(), 1, models.ItemKindBook, "El imperio final", "", "OL2W", "", cover)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, cover, item.CoverURL)
	assert.Equal(t, models.WorkDetails{WorkKey: "OL2W", Title: "The Final Empire", Description: "Ash falls.", CoverURL: cover}, saved)
}

func TestAddItemUC_Execute_CatalogDown_StillAddsTheBook(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := newUC(&mocks.MockItemRepository{}, &mocks.MockWorkRepository{}, &mocks.MockCatalogLookup{
		BooksFn: func(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
			return nil, errors.New("timeout")
		},
	})

	// Act
	item, err := uc.Execute(context.Background(), 1, models.ItemKindBook, "Elantris", "", "", "", "")

	// Assert
	require.NoError(t, err)
	assert.Empty(t, item.WorkKey)
	assert.Empty(t, item.CoverURL)
}
