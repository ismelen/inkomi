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
	works, err := client.SearchWorks(context.Background(), openlibrary.SearchQuery{Title: "la nena", Author: "mola", Limit: 5})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "/search.json", got.URL.Path)
	assert.Equal(t, "la nena", got.URL.Query().Get("title"))
	assert.Equal(t, "mola", got.URL.Query().Get("author"))
	assert.Equal(t, "5", got.URL.Query().Get("limit"))
	assert.Empty(t, got.URL.Query().Get("subject"))
	assert.Equal(t, "test-agent", got.Header.Get("User-Agent"))
	assert.Equal(t, []openlibrary.Work{{
		Key: "OL1W", Title: "La nena", Authors: []string{"Carmen Mola"}, AuthorKeys: []string{"OL9A"},
		Subjects: []string{"Thriller"}, Languages: []string{"spa"}, FirstPublishYear: 2020, CoverId: 42,
	}}, works)
}

func TestClient_WorkDetail_Descriptions_ReadsBothShapes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		body string
		want string
	}{
		{"Plain", `{"description":" Una novela ","subjects":["Crime"]}`, "Una novela"},
		{"Typed", `{"description":{"type":"/type/text","value":"Typed text"}}`, "Typed text"},
		{"Missing", `{}`, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/works/OL1W.json", r.URL.Path)
				fmt.Fprint(w, tc.body)
			})

			// Act
			detail, err := client.WorkDetail(context.Background(), "/works/OL1W")

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.want, detail.Description)
		})
	}
}

func TestClient_SubjectWorks_Subject_ReadsWorksWithAuthors(t *testing.T) {
	t.Parallel()

	// Arrange
	client := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/subjects/science_fiction.json", r.URL.Path)
		fmt.Fprint(w, `{"works":[{"key":"/works/OL2W","title":"Dune","cover_id":7,"first_publish_year":1965,
			"subject":["Science fiction"],"authors":[{"key":"/authors/OL3A","name":"Frank Herbert"}]}]}`)
	})

	// Act
	works, err := client.SubjectWorks(context.Background(), "Science Fiction", 10)

	// Assert
	require.NoError(t, err)
	require.Len(t, works, 1)
	assert.Equal(t, "OL2W", works[0].Key)
	assert.Equal(t, []string{"OL3A"}, works[0].AuthorKeys)
	assert.Equal(t, []string{"Frank Herbert"}, works[0].Authors)
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
