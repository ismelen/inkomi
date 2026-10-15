package searchbooks_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/library-service/internal/usecases/searchbooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSearchBooksUC_Execute_EmptyQuery_ReturnsErrEmptyQuery(t *testing.T) {
	t.Parallel()

	// Arrange
	called := false
	uc := searchbooks.NewSearchBooksUC(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string) ([]models.Book, error) {
			called = true
			return nil, nil
		},
	})

	// Act
	books, err := uc.Execute(context.Background(), "   ", "", nil)

	// Assert
	assert.ErrorIs(t, err, models.ErrEmptyQuery)
	assert.Nil(t, books)
	assert.False(t, called)
}

func TestSearchBooksUC_Execute_LibraryFails_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := searchbooks.NewSearchBooksUC(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string) ([]models.Book, error) {
			return nil, models.ErrNoMirror
		},
	})

	// Act
	_, err := uc.Execute(context.Background(), "dune", "", nil)

	// Assert
	assert.True(t, errors.Is(err, models.ErrNoMirror))
}

func TestSearchBooksUC_Execute_Results_TrimsQueryAndFilters(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotQuery string
	uc := searchbooks.NewSearchBooksUC(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string) ([]models.Book, error) {
			gotQuery = query
			return []models.Book{
				{Title: "Dune", Language: "English", Extension: "epub", MD5: "a"},
				{Title: "Dune", Language: "English", Extension: "epub", MD5: "b"},
				{Title: "Dune Messiah", Language: "English", Extension: "pdf", MD5: "c"},
				{Title: "Duna", Language: "Spanish", Extension: "epub", MD5: "d"},
			}, nil
		},
	})

	// Act
	books, err := uc.Execute(context.Background(), "  dune ", "english", []string{"epub"})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "dune", gotQuery)
	assert.Equal(t, []models.Book{{Title: "Dune", Language: "English", Extension: "epub", MD5: "a"}}, books)
}
