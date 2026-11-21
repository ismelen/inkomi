package wikimedia_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/platform/wikimedia"
)

func newClient(t *testing.T, handler http.HandlerFunc) *wikimedia.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return wikimedia.NewClient(server.URL+"/sparql", server.URL+"/{lang}", "test-agent")
}

func TestClient_ArticleFor_Match_ReturnsTheDecodedTitle(t *testing.T) {
	t.Parallel()

	// Arrange
	var got *http.Request
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		got = r
		fmt.Fprint(w, `{"results":{"bindings":[{"article":{"type":"uri",
			"value":"https://es.wikipedia.org/wiki/El_imperio_final_(novela)%3F"}}]}}`)
	})

	// Act
	title, err := client.ArticleFor(context.Background(), []string{"OL5W", "", "OL1M"}, "es")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "El_imperio_final_(novela)?", title)
	assert.Equal(t, "/sparql", got.URL.Path)
	assert.Equal(t, "test-agent", got.Header.Get("User-Agent"))
	query := got.URL.Query().Get("query")
	assert.Contains(t, query, `VALUES ?id { "OL5W" "OL1M" }`)
	assert.Contains(t, query, "<https://es.wikipedia.org/>")
}

func TestClient_ArticleFor_NoMatchOrNoKeys_ReturnsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	calls := 0
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		fmt.Fprint(w, `{"results":{"bindings":[]}}`)
	})

	// Act
	title, err := client.ArticleFor(context.Background(), []string{"OL5W"}, "es")
	none, noneErr := client.ArticleFor(context.Background(), nil, "es")

	// Assert
	require.NoError(t, err)
	require.NoError(t, noneErr)
	assert.Empty(t, title)
	assert.Empty(t, none)
	assert.Equal(t, 1, calls)
}

func TestClient_Summary_Kinds_ReturnsOnlyStandardExtracts(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		body   string
		want   string
	}{
		{"Standard", http.StatusOK, `{"type":"standard","extract":" Una novela. "}`, "Una novela."},
		{"Disambiguation", http.StatusOK, `{"type":"disambiguation","extract":"Puede referirse a"}`, ""},
		{"Missing", http.StatusNotFound, `{}`, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var path string
			client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
				path = r.URL.EscapedPath()
				w.WriteHeader(tc.status)
				fmt.Fprint(w, tc.body)
			})

			// Act
			text, err := client.Summary(context.Background(), "es", "El imperio final")

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.want, text)
			assert.Equal(t, "/es/api/rest_v1/page/summary/El_imperio_final", path)
		})
	}
}

func TestClient_Summary_ServerError_ReturnsIt(t *testing.T) {
	t.Parallel()

	// Arrange
	client := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	// Act
	_, err := client.Summary(context.Background(), "es", "X")

	// Assert
	assert.ErrorContains(t, err, "HTTP 503")
}
