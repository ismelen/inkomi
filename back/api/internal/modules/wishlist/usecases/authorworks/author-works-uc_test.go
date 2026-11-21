package authorworks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/authorworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
)

func TestAuthorWorksUC_Execute_InvalidQuery_ReturnsErrorWithoutLookup(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		query models.AuthorWorksQuery
		err   error
	}{
		{"No author", models.AuthorWorksQuery{Author: "  ", Page: 1}, models.ErrEmptyQuery},
		{"Invalid sort", models.AuthorWorksQuery{AuthorKey: "OL1A", Sort: "rating", Page: 1}, models.ErrInvalidSort},
		{"Zero page", models.AuthorWorksQuery{AuthorKey: "OL1A"}, models.ErrInvalidPage},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			called := false
			uc := authorworks.NewAuthorWorksUC(&mocks.MockCatalogLookup{
				WorksFn: func(ctx context.Context, query models.AuthorWorksQuery, limit int) (models.WorksPage, error) {
					called = true
					return models.WorksPage{}, nil
				},
			})

			// Act
			_, err := uc.Execute(context.Background(), tc.query)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.False(t, called)
		})
	}
}

func TestAuthorWorksUC_Execute_ValidQuery_TrimsAndDefaultsSort(t *testing.T) {
	t.Parallel()

	// Arrange
	var got models.AuthorWorksQuery
	var gotLimit int
	uc := authorworks.NewAuthorWorksUC(&mocks.MockCatalogLookup{
		WorksFn: func(ctx context.Context, query models.AuthorWorksQuery, limit int) (models.WorksPage, error) {
			got, gotLimit = query, limit
			return models.WorksPage{Works: []models.Suggestion{{Title: "Elantris"}}, Total: 30, Page: 2}, nil
		},
	})

	// Act
	page, err := uc.Execute(context.Background(), models.AuthorWorksQuery{Author: " Brandon Sanderson ", Language: "es ", Page: 2})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, models.AuthorWorksQuery{Author: "Brandon Sanderson", Sort: models.WorksSortPopular, Language: "es", Page: 2}, got)
	assert.Equal(t, 20, gotLimit)
	require.Len(t, page.Works, 1)
	assert.Equal(t, "Elantris", page.Works[0].Title)
}

func TestAuthorWorksUC_Execute_CatalogFails_ReturnsErrLookupFailure(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := authorworks.NewAuthorWorksUC(&mocks.MockCatalogLookup{
		WorksFn: func(ctx context.Context, query models.AuthorWorksQuery, limit int) (models.WorksPage, error) {
			return models.WorksPage{}, errors.New("timeout")
		},
	})

	// Act
	_, err := uc.Execute(context.Background(), models.AuthorWorksQuery{AuthorKey: "OL1A", Sort: models.WorksSortNew, Page: 1})

	// Assert
	assert.ErrorIs(t, err, models.ErrLookupFailure)
}
