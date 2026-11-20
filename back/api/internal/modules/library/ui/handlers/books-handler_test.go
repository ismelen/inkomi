package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/getcover"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/searchbooks"
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
		SearchFn: func(ctx context.Context, query string, field models.SearchField) ([]models.Book, error) {
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
	assert.Equal(t, "/books/cover?url=http%3A%2F%2Fc%2F1.jpg", body[0]["coverUrl"])
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
		{"Invalid field", "/books/search?q=dune&by=isbn", nil, http.StatusBadRequest},
		{"No mirror", "/books/search?q=dune", models.ErrNoMirror, http.StatusServiceUnavailable},
		{"Mirror error", "/books/search?q=dune", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			router := newRouter(&mocks.MockLibrary{
				SearchFn: func(ctx context.Context, query string, field models.SearchField) ([]models.Book, error) {
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

func TestBooksHandler_HandleSearch_ByAuthor_SearchesAuthor(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotField models.SearchField
	router := newRouter(&mocks.MockLibrary{
		SearchFn: func(ctx context.Context, query string, field models.SearchField) ([]models.Book, error) {
			gotField = field
			return nil, nil
		},
	})
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/books/search?q=mola&by=author", nil))

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, models.SearchFieldAuthor, gotField)
}

func TestCoversHandler_HandleCover_Results_MapsStatusAndStreams(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		err    error
		status int
	}{
		{"Image", nil, http.StatusOK},
		{"Not a cover", models.ErrInvalidCoverURL, http.StatusNotFound},
		{"Mirror down", errors.New("boom"), http.StatusBadGateway},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			router := chi.NewRouter()
			routes.SetupCoverRoutes(router, handlers.NewCoversHandler(getcover.NewGetCoverUC(&mocks.MockCoverSource{
				CoverFn: func(ctx context.Context, url string) (*models.Cover, error) {
					if tc.err != nil {
						return nil, tc.err
					}
					return &models.Cover{Body: io.NopCloser(strings.NewReader("jpg")), ContentType: "image/jpeg"}, nil
				},
			})))
			rec := httptest.NewRecorder()

			// Act
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/books/cover?url=https%3A%2F%2Flibgen.la%2Fcovers%2Fa.jpg", nil))

			// Assert
			assert.Equal(t, tc.status, rec.Code)
			if tc.err == nil {
				assert.Equal(t, "image/jpeg", rec.Header().Get("Content-Type"))
				assert.Equal(t, "jpg", rec.Body.String())
			}
		})
	}
}
