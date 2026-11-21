package searchbooks_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/searchbooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func epubs(prefix string, n int) []models.Book {
	books := make([]models.Book, n)
	for i := range books {
		books[i] = models.Book{Title: fmt.Sprintf("%s %d", prefix, i), Language: "English", Extension: "epub"}
	}
	return books
}

func TestSearchBooksUC_Execute_InvalidInput_ReturnsErrorWithoutSearching(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		query string
		field models.SearchField
		page  int
		err   error
	}{
		{"Empty query", "   ", models.SearchFieldTitle, 1, models.ErrEmptyQuery},
		{"Invalid field", "mola", models.SearchField("isbn"), 1, models.ErrInvalidField},
		{"Zero page", "mola", models.SearchFieldTitle, 0, models.ErrInvalidPage},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			called := false
			uc := searchbooks.NewSearchBooksUC(&mocks.MockLibrary{
				SearchFn: func(ctx context.Context, query string, field models.SearchField, page int) (models.SearchPage, error) {
					called = true
					return models.SearchPage{}, nil
				},
			})

			// Act
			result, err := uc.Execute(context.Background(), tc.query, tc.field, "", nil, tc.page)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Empty(t, result.Books)
			assert.False(t, called)
		})
	}
}

func TestSearchBooksUC_Execute_LibraryFails_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := searchbooks.NewSearchBooksUC(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string, field models.SearchField, page int) (models.SearchPage, error) {
			return models.SearchPage{}, models.ErrNoMirror
		},
	})

	// Act
	_, err := uc.Execute(context.Background(), "dune", models.SearchFieldTitle, "", nil, 1)

	// Assert
	assert.True(t, errors.Is(err, models.ErrNoMirror))
}

func TestSearchBooksUC_Execute_Results_TrimsQueryAndFilters(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotQuery string
	var gotField models.SearchField
	uc := searchbooks.NewSearchBooksUC(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string, field models.SearchField, page int) (models.SearchPage, error) {
			gotQuery, gotField = query, field
			return models.SearchPage{Books: []models.Book{
				{Title: "Dune", Language: "English", Extension: "epub", MD5: "a"},
				{Title: "Dune", Language: "English", Extension: "epub", MD5: "b"},
				{Title: "Dune Messiah", Language: "English", Extension: "pdf", MD5: "c"},
				{Title: "Duna", Language: "Spanish", Extension: "epub", MD5: "d"},
			}}, nil
		},
	})

	// Act
	result, err := uc.Execute(context.Background(), "  dune ", models.SearchFieldAuthor, "english", []string{"epub"}, 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "dune", gotQuery)
	assert.Equal(t, models.SearchFieldAuthor, gotField)
	assert.Equal(t, []models.Book{{Title: "Dune", Language: "English", Extension: "epub", MD5: "a"}}, result.Books)
	assert.Zero(t, result.NextPage)
}

func TestSearchBooksUC_Execute_FewFilteredBooks_ReadsNextPages(t *testing.T) {
	t.Parallel()

	// Arrange
	var pages []int
	uc := searchbooks.NewSearchBooksUC(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string, field models.SearchField, page int) (models.SearchPage, error) {
			pages = append(pages, page)
			return models.SearchPage{Books: epubs(fmt.Sprint("page ", page), 4), NextPage: page + 4}, nil
		},
	})

	// Act
	result, err := uc.Execute(context.Background(), "sanderson", models.SearchFieldAuthor, "", []string{"epub"}, 5)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []int{5, 9}, pages)
	assert.Len(t, result.Books, 8)
	assert.Equal(t, 13, result.NextPage)
}

func TestSearchBooksUC_Execute_EnoughBooks_StopsAfterOnePage(t *testing.T) {
	t.Parallel()

	// Arrange
	calls := 0
	uc := searchbooks.NewSearchBooksUC(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string, field models.SearchField, page int) (models.SearchPage, error) {
			calls++
			return models.SearchPage{Books: epubs("dune", 10), NextPage: 5}, nil
		},
	})

	// Act
	result, err := uc.Execute(context.Background(), "dune", models.SearchFieldTitle, "", nil, 1)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, calls)
	assert.Len(t, result.Books, 10)
	assert.Equal(t, 5, result.NextPage)
}

func TestSearchBooksUC_Execute_LaterPageFails_ReturnsWhatItHas(t *testing.T) {
	t.Parallel()

	// Arrange
	uc := searchbooks.NewSearchBooksUC(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string, field models.SearchField, page int) (models.SearchPage, error) {
			if page > 1 {
				return models.SearchPage{}, errors.New("captcha")
			}
			return models.SearchPage{Books: epubs("dune", 2), NextPage: 5}, nil
		},
	})

	// Act
	result, err := uc.Execute(context.Background(), "dune", models.SearchFieldTitle, "", nil, 1)

	// Assert
	require.NoError(t, err)
	assert.Len(t, result.Books, 2)
	assert.Equal(t, 5, result.NextPage, "the failed page is asked for again")
}
