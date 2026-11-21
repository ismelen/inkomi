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

func TestClient_GetWork_Record_ReadsTypedDescriptionAndSkipsBadCovers(t *testing.T) {
	t.Parallel()

	// Arrange
	var path string
	client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		fmt.Fprint(w, `{"key":"/works/OL2W","title":"Mistborn","description":{"type":"/type/text","value":"Ash falls."},
			"covers":[-1,42],"subjects":["Fantasy"],"first_publish_date":"2006"}`)
	})

	// Act
	work, err := client.GetWork(context.Background(), "OL2W")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "/works/OL2W.json", path)
	assert.Equal(t, openlibrary.WorkRecord{
		Key: "OL2W", Title: "Mistborn", Description: "Ash falls.", CoverIds: []int{42},
		Subjects: []string{"Fantasy"}, FirstPublishDate: "2006",
	}, work)
}

func TestClient_GetWork_PlainDescriptionAndMissing_MapsThem(t *testing.T) {
	t.Parallel()

	// Arrange
	client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/works/OL404W.json" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		fmt.Fprint(w, `{"key":"/works/OL3W","title":"Elantris","description":"A city."}`)
	})

	// Act
	work, err := client.GetWork(context.Background(), "OL3W")
	_, missingErr := client.GetWork(context.Background(), "OL404W")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "A city.", work.Description)
	assert.ErrorIs(t, missingErr, openlibrary.ErrNotFound)
}

func TestClient_EditionKeyIn_EditionInLanguage_ReturnsItsKey(t *testing.T) {
	t.Parallel()

	// Arrange
	var got *http.Request
	client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		got = r
		fmt.Fprint(w, `{"docs":[{"key":"/works/OL5W","editions":{"docs":[
			{"key":"/books/OL29659236M","language":["spa"]}]}}]}`)
	})

	// Act
	key, err := client.EditionKeyIn(context.Background(), "OL5W", "es")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "OL29659236M", key)
	assert.Equal(t, "key:/works/OL5W", got.URL.Query().Get("q"))
	assert.Equal(t, "spa", got.URL.Query().Get("language"))
}

func TestClient_EditionKeyIn_OtherLanguageOrUnknown_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	calls := 0
	client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprint(w, `{"docs":[{"editions":{"docs":[{"key":"/books/OL1M","language":["eng"]}]}}]}`)
	})

	// Act
	key, err := client.EditionKeyIn(context.Background(), "OL5W", "es")
	unknown, unknownErr := client.EditionKeyIn(context.Background(), "OL5W", "xx")

	// Assert
	require.NoError(t, err)
	require.NoError(t, unknownErr)
	assert.Empty(t, key)
	assert.Empty(t, unknown)
	assert.Equal(t, 1, calls)
}
