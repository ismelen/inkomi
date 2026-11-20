package lookup_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/lookup"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
)

func TestLookupUC_Execute_Kinds_AsksTheirCatalog(t *testing.T) {
	t.Parallel()

	// Arrange
	var asked []string
	uc := lookup.NewLookupUC(&mocks.MockCatalogLookup{
		BooksFn: func(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
			asked = append(asked, "books:"+query)
			return []models.Suggestion{{Title: "La nena"}}, nil
		},
		AuthorsFn: func(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
			asked = append(asked, "authors:"+query)
			return nil, nil
		},
	})

	// Act
	books, err := uc.Execute(context.Background(), models.ItemKindBook, " la nena ")
	require.NoError(t, err)
	_, err = uc.Execute(context.Background(), models.ItemKindAuthor, "mola")

	// Assert
	require.NoError(t, err)
	assert.Len(t, books, 1)
	assert.Equal(t, []string{"books:la nena", "authors:mola"}, asked)
}

func TestLookupUC_Execute_Failures_MapsErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		kind  models.ItemKind
		query string
		err   error
	}{
		{"Invalid kind", models.ItemKind("movie"), "dune", models.ErrInvalidKind},
		{"Short query", models.ItemKindBook, " a ", models.ErrEmptyQuery},
		{"Catalog down", models.ItemKindBook, "dune", models.ErrLookupFailure},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := lookup.NewLookupUC(&mocks.MockCatalogLookup{
				BooksFn: func(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
					return nil, errors.New("HTTP 503")
				},
			})

			// Act
			_, err := uc.Execute(context.Background(), tc.kind, tc.query)

			// Assert
			assert.ErrorIs(t, err, tc.err)
		})
	}
}
