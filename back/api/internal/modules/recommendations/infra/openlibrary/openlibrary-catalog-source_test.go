package openlibrary_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	recsol "github.com/ismelen/inkomi/back/api/internal/modules/recommendations/infra/openlibrary"
	"github.com/ismelen/inkomi/back/api/internal/platform/openlibrary"
)

func newSource(t *testing.T, handler http.HandlerFunc) *recsol.OpenLibraryCatalogSource {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return recsol.NewOpenLibraryCatalogSource(openlibrary.NewClient(server.URL, "test", 1000))
}

const searchResults = `{"docs":[
	{"key":"/works/OL5W","title":"La novia gitana: guía de lectura","author_name":["Otro"],"author_key":["OL2A"]},
	{"key":"/works/OL1W","title":"La novia gitana","author_name":["Carmen Mola"],"author_key":["OL9A"],"subject":["Thriller"],"first_publish_year":2018}
]}`

func TestOpenLibraryCatalogSource_MatchBook_Titles_KeepsTheClosest(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		title   string
		author  string
		wantKey string
		param   string
	}{
		{"Free text", "Carmen Mola - La novia gitana", "", "OL1W", "q"},
		{"With author", "La novia gitana", "Carmen Mola", "OL1W", "title"},
		{"Unrelated", "Dune", "", "", "q"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			source := newSource(t, func(w http.ResponseWriter, r *http.Request) {
				assert.NotEmpty(t, r.URL.Query().Get(tc.param))
				fmt.Fprint(w, searchResults)
			})

			// Act
			work, err := source.MatchBook(context.Background(), tc.title, tc.author)

			// Assert
			require.NoError(t, err)
			if tc.wantKey == "" {
				assert.Nil(t, work)
				return
			}
			require.NotNil(t, work)
			assert.Equal(t, tc.wantKey, work.Key)
			assert.Equal(t, []string{"OL9A"}, work.AuthorKeys)
		})
	}
}

func TestOpenLibraryCatalogSource_MatchAuthor_Names_MatchesSimilarOnly(t *testing.T) {
	t.Parallel()

	// Arrange
	source := newSource(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"docs":[{"key":"OL9A","name":"Carmen Mola"}]}`)
	})

	// Act
	key, name, err := source.MatchAuthor(context.Background(), "carmen mola")
	require.NoError(t, err)
	missKey, _, err := source.MatchAuthor(context.Background(), "Frank Herbert")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "OL9A", key)
	assert.Equal(t, "Carmen Mola", name)
	assert.Empty(t, missKey)
}

func TestOpenLibraryCatalogSource_NewWorks_Subject_AsksRecentWorks(t *testing.T) {
	t.Parallel()

	// Arrange
	var query map[string][]string
	source := newSource(t, func(w http.ResponseWriter, r *http.Request) {
		query = r.URL.Query()
		fmt.Fprint(w, searchResults)
	})

	// Act
	works, err := source.NewWorks(context.Background(), "Thriller", "", 2025, 20)

	// Assert
	require.NoError(t, err)
	assert.Len(t, works, 2)
	assert.Equal(t, "Thriller", query["subject"][0])
	assert.Equal(t, "new", query["sort"][0])
	assert.Equal(t, "first_publish_year:[2025 TO *]", query["q"][0])
}

func TestOpenLibraryCatalogSource_DescriptionAndSubject_NotFound_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	source := newSource(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	// Act
	description, err := source.Description(context.Background(), "OL1W")
	require.NoError(t, err)
	works, err := source.WorksBySubject(context.Background(), "nothing", 10)

	// Assert
	require.NoError(t, err)
	assert.Empty(t, description)
	assert.Empty(t, works)
}
