package getwork_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/getwork"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
)

func TestGetWorkUC_Execute_Stored_ReturnsItWithoutCatalog(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := getwork.NewGetWorkUC(&mocks.MockWorkRepository{
		GetFn: func(ctx context.Context, workKey string) (*models.WorkDetails, error) {
			return &models.WorkDetails{WorkKey: workKey, Description: "Stored."}, nil
		},
	}, &mocks.MockCatalogLookup{
		WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
			t.Fatal("the catalog isn't asked for a stored work")
			return models.WorkDetails{}, nil
		},
	})

	// Act
	work, err := uc.Execute(context.Background(), "OL5W")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Stored.", work.Description)
}

func TestGetWorkUC_Execute_NotStored_FetchesAndSaves(t *testing.T) {
	t.Parallel()

	// Arrange
	var saved models.WorkDetails
	uc := getwork.NewGetWorkUC(&mocks.MockWorkRepository{
		SaveFn: func(ctx context.Context, work models.WorkDetails) error {
			saved = work
			return nil
		},
	}, &mocks.MockCatalogLookup{
		WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
			return models.WorkDetails{Title: "Mistborn", Description: "Ash falls."}, nil
		},
	})

	// Act
	work, err := uc.Execute(context.Background(), "OL5W")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Ash falls.", work.Description)
	assert.Equal(t, models.WorkDetails{WorkKey: "OL5W", Title: "Mistborn", Description: "Ash falls."}, saved)
}

func TestGetWorkUC_Execute_Failures_MapsThem(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		workKey string
		catalog error
		err     error
	}{
		{"Invalid key", "../x", nil, models.ErrInvalidWorkKey},
		{"Author key", "OL1A", nil, models.ErrInvalidWorkKey},
		{"Unknown", "OL5W", models.ErrWorkNotFound, models.ErrWorkNotFound},
		{"Catalog down", "OL5W", errors.New("timeout"), models.ErrLookupFailure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := getwork.NewGetWorkUC(&mocks.MockWorkRepository{}, &mocks.MockCatalogLookup{
				WorkFn: func(ctx context.Context, workKey string) (models.WorkDetails, error) {
					return models.WorkDetails{}, tc.catalog
				},
			})

			// Act
			_, err := uc.Execute(context.Background(), tc.workKey)

			// Assert
			assert.ErrorIs(t, err, tc.err)
		})
	}
}
