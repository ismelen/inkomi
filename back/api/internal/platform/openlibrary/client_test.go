package openlibrary_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/platform/openlibrary"
)

func newServer(t *testing.T, handler http.HandlerFunc) *openlibrary.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return openlibrary.NewClient(server.URL, "test-agent", 1000)
}

func TestClient_SearchWorks_Query_SendsParamsAndTrimsKeys(t *testing.T) {
	t.Parallel()

	// Arrange
	var got *http.Request
	client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = r
		fmt.Fprint(w, `{"docs":[{"key":"/works/OL1W","title":"La nena","author_name":["Carmen Mola"],"author_key":["OL9A"],
			"subject":["Thriller"],"first_publish_year":2020,"language":["spa"],"cover_i":42},{"key":"","title":"no key"}]}`)
	})

	// Act
	result, err := client.SearchWorks(context.Background(), openlibrary.SearchQuery{Title: "la nena", Author: "mola", Limit: 5})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "/search.json", got.URL.Path)
	assert.Equal(t, "la nena", got.URL.Query().Get("title"))
	assert.Equal(t, "mola", got.URL.Query().Get("author"))
	assert.Equal(t, "5", got.URL.Query().Get("limit"))
	assert.Empty(t, got.URL.Query().Get("subject"))
	assert.Empty(t, got.URL.Query().Get("page"))
	assert.Empty(t, got.URL.Query().Get("language"))
	assert.Equal(t, "test-agent", got.Header.Get("User-Agent"))
	assert.Equal(t, []openlibrary.Work{{
		Key: "OL1W", Title: "La nena", Authors: []string{"Carmen Mola"}, AuthorKeys: []string{"OL9A"},
		Subjects: []string{"Thriller"}, Languages: []string{"spa"}, FirstPublishYear: 2020, CoverId: 42,
	}}, result.Works)
}

func TestClient_SearchAuthors_NotFoundAndErrors_MapsThem(t *testing.T) {
	t.Parallel()

	// Arrange
	client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	// Act
	_, err := client.SearchAuthors(context.Background(), "mola", 5)

	// Assert
	assert.ErrorIs(t, err, openlibrary.ErrNotFound)
}

func TestCoverURL_CoverId_BuildsMediumCover(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "https://covers.openlibrary.org/b/id/42-M.jpg", openlibrary.CoverURL(42))
	assert.Empty(t, openlibrary.CoverURL(0))
}

func TestClient_SearchWorks_AuthorPageWithLanguage_ReadsEditionAndTotal(t *testing.T) {
	t.Parallel()

	// Arrange
	var got *http.Request
	client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = r
		fmt.Fprint(w, `{"numFound":44,"docs":[{"key":"/works/OL2W","title":"The Final Empire","cover_i":1,
			"editions":{"docs":[{"title":"El imperio final","cover_i":7}]}}]}`)
	})

	// Act
	result, err := client.SearchWorks(context.Background(), openlibrary.SearchQuery{
		AuthorKey: "OL1A", Sort: "new", Language: "es", Limit: 20, Page: 3,
	})

	// Assert
	require.NoError(t, err)
	query := got.URL.Query()
	assert.Equal(t, "OL1A", query.Get("author_key"))
	assert.Equal(t, "new", query.Get("sort"))
	assert.Equal(t, "spa", query.Get("language"))
	assert.Equal(t, "es", query.Get("lang"))
	assert.Equal(t, "3", query.Get("page"))
	assert.Contains(t, query.Get("fields"), "editions.title")
	assert.Equal(t, 44, result.Total)
	require.Len(t, result.Works, 1)
	assert.Equal(t, "El imperio final", result.Works[0].Title)
	assert.Equal(t, 7, result.Works[0].CoverId)
}
