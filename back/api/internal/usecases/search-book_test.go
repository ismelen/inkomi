package usecases

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/test/mocks"
)

func newSearchUC(hasMirror bool, books []book.Book, searchErr error) (*SearchBookUC, *mocks.BooksProviderMock) {
	src := mocks.NewBooksSourceMock("http://test.example")
	src.SearchResult = books
	src.SearchErr = searchErr

	prov := &mocks.BooksProviderMock{
		Source:    src,
		HasMirror: hasMirror,
	}
	return NewSearchBookUC(prov), prov
}

func TestSearchBookUC_Execute_NoMirror_ReturnsError(t *testing.T) {
	t.Parallel()
	// Arrange
	uc, _ := newSearchUC(false, nil, nil)

	// Act
	_, err := uc.Execute(context.Background(), "golang", "English", []string{"epub"})

	// Assert
	require.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "mirror")
}

func TestSearchBookUC_Execute_FiltersByLanguage_ShouldFilter(t *testing.T) {
	t.Parallel()
	// Arrange
	books := []book.Book{
		{Title: "Go Programming", Language: "English", Extension: "epub", MD5: "aaa111"},
		{Title: "Programacion Go", Language: "Spanish", Extension: "epub", MD5: "bbb222"},
	}
	uc, _ := newSearchUC(true, books, nil)

	// Act
	result, err := uc.Execute(context.Background(), "go", "English", []string{"epub"})

	// Assert
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "English", result[0].Language)
}

func TestSearchBookUC_Execute_FiltersByFormat_ShouldFilter(t *testing.T) {
	t.Parallel()
	// Arrange
	books := []book.Book{
		{Title: "Go Programming", Language: "English", Extension: "epub", MD5: "aaa111"},
		{Title: "Go Reference", Language: "English", Extension: "pdf", MD5: "bbb222"},
	}
	uc, _ := newSearchUC(true, books, nil)

	// Act
	result, err := uc.Execute(context.Background(), "go", "English", []string{"epub"})

	// Assert
	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "epub", result[0].Extension)
}

func TestSearchBookUC_Execute_Deduplicates_ShouldDeduplicate(t *testing.T) {
	t.Parallel()
	// Arrange
	// Two books with the same title (after normalization) and same MD5 should collapse to one.
	books := []book.Book{
		{Title: "Go Programming", Language: "English", Extension: "epub", MD5: "aaa111"},
		{Title: "Go Programming", Language: "English", Extension: "epub", MD5: "aaa111"},
	}
	uc, _ := newSearchUC(true, books, nil)

	// Act
	result, err := uc.Execute(context.Background(), "go", "English", []string{"epub"})

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestSearchBookUC_Execute_MirrorError_ShouldPropagate(t *testing.T) {
	t.Parallel()
	// Arrange
	searchErr := errors.New("mirror down")
	uc, _ := newSearchUC(true, nil, searchErr)

	// Act
	_, err := uc.Execute(context.Background(), "go", "", []string{"epub"})

	// Assert
	require.ErrorIs(t, err, searchErr)
}

func TestSearchBookUC_Execute_EmptyLanguage_ReturnsAll(t *testing.T) {
	t.Parallel()
	// Arrange
	books := []book.Book{
		{Title: "Go Programming", Language: "English", Extension: "epub", MD5: "aaa111"},
		{Title: "Programacion Go", Language: "Spanish", Extension: "pdf", MD5: "bbb222"},
	}
	uc, _ := newSearchUC(true, books, nil)

	// Act
	// language="" -> LanguageFilter passes through; formats=["epub","pdf"] -> both pass
	result, err := uc.Execute(context.Background(), "go", "", []string{"epub", "pdf"})

	// Assert
	require.NoError(t, err)
	assert.Len(t, result, 2)
}
