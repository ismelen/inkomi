package wikipedia_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/wikipedia"
	"github.com/ismelen/inkomi/back/api/internal/platform/openlibrary"
	"github.com/ismelen/inkomi/back/api/internal/platform/wikimedia"
)

type fakeWiki struct {
	mu      sync.Mutex
	queries []string
}

func (f *fakeWiki) lastQuery() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.queries) == 0 {
		return ""
	}
	return f.queries[len(f.queries)-1]
}

// newLookup serves edition as the Open Library search (or fails it with a 500 when empty),
// article as the Wikidata match and summary as the article lead.
func newLookup(t *testing.T, edition, article, summary string) (*wikipedia.WikipediaSynopsisLookup, *fakeWiki) {
	t.Helper()

	ol := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if edition == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, edition)
	}))
	t.Cleanup(ol.Close)
	fake := &fakeWiki{}
	wiki := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/sparql") {
			fake.mu.Lock()
			fake.queries = append(fake.queries, r.URL.Query().Get("query"))
			fake.mu.Unlock()
			bindings := ""
			if article != "" {
				bindings = `{"article":{"value":"https://es.wikipedia.org/wiki/` + article + `"}}`
			}
			fmt.Fprint(w, `{"results":{"bindings":[`+bindings+`]}}`)
			return
		}
		fmt.Fprint(w, summary)
	}))
	t.Cleanup(wiki.Close)
	return wikipedia.NewWikipediaSynopsisLookup(
		openlibrary.NewClient(ol.URL, "test", 1000),
		wikimedia.NewClient(wiki.URL+"/sparql", wiki.URL+"/{lang}", "test"),
	), fake
}

const spanishEdition = `{"docs":[{"editions":{"docs":[{"key":"/books/OL29659236M","language":["spa"]}]}}]}`

func TestWikipediaSynopsisLookup_Synopsis_Article_ReturnsItsLead(t *testing.T) {
	t.Parallel()

	// Arrange
	lookup, fake := newLookup(t, spanishEdition, "El_camino_de_los_reyes",
		`{"type":"standard","extract":"El camino de los reyes es una novela."}`)

	// Act
	text, err := lookup.Synopsis(context.Background(), "OL5W", "es")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "El camino de los reyes es una novela.", text)
	assert.Contains(t, fake.lastQuery(), `"OL5W" "OL29659236M"`)
}

func TestWikipediaSynopsisLookup_Synopsis_NoArticle_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	lookup, _ := newLookup(t, `{"docs":[]}`, "", `{"type":"standard","extract":"never read"}`)

	// Act
	text, err := lookup.Synopsis(context.Background(), "OL5W", "es")

	// Assert
	require.NoError(t, err)
	assert.Empty(t, text)
}

func TestWikipediaSynopsisLookup_Synopsis_EditionLookupFails_UsesTheWorkKey(t *testing.T) {
	t.Parallel()

	// Arrange
	lookup, fake := newLookup(t, "", "Dune", `{"type":"standard","extract":"Dune es una novela."}`)

	// Act
	text, err := lookup.Synopsis(context.Background(), "OL5W", "es")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Dune es una novela.", text)
	assert.Contains(t, fake.lastQuery(), `VALUES ?id { "OL5W" }`)
}
