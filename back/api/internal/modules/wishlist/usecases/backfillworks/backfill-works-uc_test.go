package backfillworks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/backfillworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/resolvework"
)

func TestBackfillWorksUC_Execute_PendingBooks_ResolvesAndCountsFailures(t *testing.T) {
	t.Parallel()

	// Arrange
	items := &mocks.MockItemRepository{
		ListWithoutFn: func(ctx context.Context) ([]models.Item, error) {
			return []models.Item{
				{Id: 1, Kind: models.ItemKindBook, Title: "Mistborn", WorkKey: "OL5W"},
				{Id: 2, Kind: models.ItemKindBook, Title: "Down", WorkKey: "OL6W"},
			}, nil
		},
	}
	catalog := &mocks.MockCatalogLookup{
		WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
			if workKey == "OL6W" {
				return models.WorkDetails{}, errors.New("timeout")
			}
			return models.WorkDetails{CoverURL: "https://covers.openlibrary.org/b/id/1-M.jpg"}, nil
		},
	}
	uc := backfillworks.NewBackfillWorksUC(items, resolvework.NewResolveWorkUC(items, &mocks.MockWorkRepository{}, catalog))

	// Act
	resolved, failed, err := uc.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, resolved)
	assert.Equal(t, 1, failed)
}

func TestBackfillWorksUC_Execute_ListFails_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	items := &mocks.MockItemRepository{
		ListWithoutFn: func(ctx context.Context) ([]models.Item, error) { return nil, errors.New("db") },
	}
	uc := backfillworks.NewBackfillWorksUC(items, resolvework.NewResolveWorkUC(items, &mocks.MockWorkRepository{}, &mocks.MockCatalogLookup{}))

	// Act
	_, _, err := uc.Execute(context.Background())

	// Assert
	assert.Error(t, err)
}
