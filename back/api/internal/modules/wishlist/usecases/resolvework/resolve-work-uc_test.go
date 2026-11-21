package resolvework_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/resolvework"
)

const (
	workCover  = "https://covers.openlibrary.org/b/id/1-M.jpg"
	matchCover = "https://covers.openlibrary.org/b/id/2-M.jpg"
	hintCover  = "https://covers.openlibrary.org/b/id/3-M.jpg"
)

func catalog() *mocks.MockCatalogLookup {
	return &mocks.MockCatalogLookup{
		BooksFn: func(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
			return []models.Suggestion{{WorkKey: "OL5W", CoverURL: matchCover}}, nil
		},
		WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
			return models.WorkDetails{Title: "Mistborn", Description: "Ash falls.", CoverURL: workCover}, nil
		},
	}
}

func TestResolveWorkUC_Execute_TypedBook_FindsWorkAndStoresDetails(t *testing.T) {
	t.Parallel()

	// Arrange
	var query string
	var saved models.WorkDetails
	lookup := catalog()
	books := lookup.BooksFn
	lookup.BooksFn = func(ctx context.Context, q string, limit int) ([]models.Suggestion, error) {
		query = q
		return books(ctx, q, limit)
	}
	uc := resolvework.NewResolveWorkUC(&mocks.MockItemRepository{}, &mocks.MockWorkRepository{
		SaveFn: func(ctx context.Context, work models.WorkDetails) error {
			saved = work
			return nil
		},
	}, lookup)
	item := &models.Item{Kind: models.ItemKindBook, Title: "Mistborn", Author: "Sanderson"}

	// Act
	err := uc.Execute(context.Background(), item, "")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Mistborn Sanderson", query)
	assert.Equal(t, "OL5W", item.WorkKey)
	assert.Equal(t, matchCover, item.CoverURL, "the cover of the match wins over the work's")
	assert.Equal(t, "Ash falls.", saved.Description)
	assert.Equal(t, "OL5W", saved.WorkKey)
}

func TestResolveWorkUC_Execute_CoverHint_UsesItUnlessForeign(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		hint string
		want string
	}{
		{"Catalog cover", hintCover, hintCover},
		{"Foreign cover", "https://evil.example/x.jpg", workCover},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := resolvework.NewResolveWorkUC(&mocks.MockItemRepository{}, &mocks.MockWorkRepository{}, catalog())
			item := &models.Item{Kind: models.ItemKindBook, Title: "Mistborn", WorkKey: "OL5W"}

			// Act
			err := uc.Execute(context.Background(), item, tc.hint)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.want, item.CoverURL)
		})
	}
}

func TestResolveWorkUC_Execute_StoredWork_KeepsItsCoverWithoutCatalog(t *testing.T) {
	t.Parallel()

	// Arrange
	saved := false
	uc := resolvework.NewResolveWorkUC(&mocks.MockItemRepository{}, &mocks.MockWorkRepository{
		GetFn: func(ctx context.Context, workKey string) (*models.WorkDetails, error) {
			return &models.WorkDetails{WorkKey: workKey, CoverURL: workCover}, nil
		},
		SaveFn: func(ctx context.Context, work models.WorkDetails) error {
			saved = true
			return nil
		},
	}, &mocks.MockCatalogLookup{
		WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
			t.Fatal("the catalog isn't asked for a stored work")
			return models.WorkDetails{}, nil
		},
	})
	item := &models.Item{Kind: models.ItemKindBook, Title: "Mistborn", WorkKey: "OL5W"}

	// Act
	err := uc.Execute(context.Background(), item, hintCover)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, workCover, item.CoverURL)
	assert.False(t, saved)
}

func TestResolveWorkUC_Execute_NoMatchOrAuthor_LeavesItAlone(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		item models.Item
	}{
		{"No match", models.Item{Kind: models.ItemKindBook, Title: "Unknown"}},
		{"Author", models.Item{Kind: models.ItemKindAuthor, Title: "Carmen Mola"}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := resolvework.NewResolveWorkUC(&mocks.MockItemRepository{}, &mocks.MockWorkRepository{}, &mocks.MockCatalogLookup{})
			item := tc.item

			// Act
			err := uc.Execute(context.Background(), &item, "")

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.item, item)
		})
	}
}

func TestResolveWorkUC_Execute_DuplicateWork_KeepsItemWithoutKey(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := resolvework.NewResolveWorkUC(&mocks.MockItemRepository{
		SetWorkKeyFn: func(ctx context.Context, item *models.Item, workKey string) error {
			return models.ErrDuplicate
		},
	}, &mocks.MockWorkRepository{}, catalog())
	item := &models.Item{Kind: models.ItemKindBook, Title: "Mistborn"}

	// Act
	err := uc.Execute(context.Background(), item, "")

	// Assert
	require.NoError(t, err)
	assert.Empty(t, item.WorkKey)
}

func TestResolveWorkUC_Execute_UnknownWork_StoresTitleOnly(t *testing.T) {
	t.Parallel()

	// Arrange
	var saved models.WorkDetails
	uc := resolvework.NewResolveWorkUC(&mocks.MockItemRepository{}, &mocks.MockWorkRepository{
		SaveFn: func(ctx context.Context, work models.WorkDetails) error {
			saved = work
			return nil
		},
	}, &mocks.MockCatalogLookup{})
	item := &models.Item{Kind: models.ItemKindBook, Title: "Mistborn", WorkKey: "OL5W"}

	// Act
	err := uc.Execute(context.Background(), item, "")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, models.WorkDetails{WorkKey: "OL5W", Title: "Mistborn"}, saved)
}

func TestResolveWorkUC_Execute_CatalogDown_ReturnsErrLookupFailure(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := resolvework.NewResolveWorkUC(&mocks.MockItemRepository{}, &mocks.MockWorkRepository{}, &mocks.MockCatalogLookup{
		WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
			return models.WorkDetails{}, errors.New("timeout")
		},
	})

	// Act
	err := uc.Execute(context.Background(), &models.Item{Kind: models.ItemKindBook, Title: "x", WorkKey: "OL5W"}, "")

	// Assert
	assert.ErrorIs(t, err, models.ErrLookupFailure)
}
