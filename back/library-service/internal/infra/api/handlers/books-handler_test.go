package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/library-service/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/library-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/library-service/internal/usecases/searchbooks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRouter(library *mocks.MockLibrary) http.Handler {
	router := chi.NewRouter()
	routes.SetupBookRoutes(router, handlers.NewBooksHandler(searchbooks.NewSearchBooksUC(library)))
	return router
}

func TestBooksHandler_HandleSearch_Results_ReturnsCamelCaseBooks(t *testing.T) {
	t.Parallel()

	// Arrange
	router := newRouter(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string) ([]models.Book, error) {
			return []models.Book{
				{Title: "Dune", Extension: "epub", MD5: "abc", CoverURL: "http://c/1.jpg"},
				{Title: "Dune Messiah", Extension: "pdf", MD5: "def"},
			}, nil
		},
	})
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/books/search?q=dune", nil))

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	var body []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body, 1, "only epub by default")
	assert.Equal(t, "Dune", body[0]["title"])
	assert.Equal(t, "abc", body[0]["md5"])
	assert.Equal(t, "http://c/1.jpg", body[0]["coverUrl"])
}

func TestBooksHandler_HandleSearch_Errors_MapsStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		url    string
		err    error
		status int
	}{
		{"Empty query", "/books/search?q=", nil, http.StatusBadRequest},
		{"No mirror", "/books/search?q=dune", models.ErrNoMirror, http.StatusServiceUnavailable},
		{"Mirror error", "/books/search?q=dune", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			router := newRouter(&mocks.MockLibrary{
				SearchFn: func(ctx context.Context, query string) ([]models.Book, error) {
					return nil, tc.err
				},
			})
			rec := httptest.NewRecorder()

			// Act
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.url, nil))

			// Assert
			assert.Equal(t, tc.status, rec.Code)
			assert.Contains(t, rec.Body.String(), `"error"`)
		})
	}
}
