package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/additem"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/authorworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/listitems"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/lookup"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/removeitem"
	"github.com/ismelen/inkomi/back/api/internal/platform/httpauth"
)

func newRouter(repo *mocks.MockItemRepository, catalog *mocks.MockCatalogLookup) http.Handler {
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpauth.WithUserClaims(r.Context(), &httpauth.UserClaims{Id: 5})))
		})
	})
	routes.SetupWishlistRoutes(router, handlers.NewWishlistHandler(
		additem.NewAddItemUC(repo), removeitem.NewRemoveItemUC(repo), listitems.NewListItemsUC(repo), lookup.NewLookupUC(catalog),
		authorworks.NewAuthorWorksUC(catalog),
	))
	return router
}

func TestWishlistHandler_HandleAdd_Body_ReturnsItemOfTheUser(t *testing.T) {
	t.Parallel()

	// Arrange
	var stored *models.Item
	router := newRouter(&mocks.MockItemRepository{
		AddFn: func(ctx context.Context, item *models.Item) error {
			stored, item.Id = item, 11
			return nil
		},
	}, &mocks.MockCatalogLookup{})
	rec := httptest.NewRecorder()
	body := `{"kind":"book","title":"La nena","author":"Carmen Mola","workKey":"OL1W"}`

	// Act
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/wishlist", strings.NewReader(body)))

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 5, stored.UserId)
	var response map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &response))
	assert.Equal(t, float64(11), response["id"])
	assert.Equal(t, "OL1W", response["workKey"])
}

func TestWishlistHandler_Errors_MapsStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		method string
		url    string
		body   string
		status int
	}{
		{"Invalid kind", http.MethodPost, "/wishlist", `{"kind":"movie","title":"x"}`, http.StatusBadRequest},
		{"Bad json", http.MethodPost, "/wishlist", `{`, http.StatusBadRequest},
		{"Duplicate", http.MethodPost, "/wishlist", `{"kind":"book","title":"dup"}`, http.StatusConflict},
		{"Bad id", http.MethodDelete, "/wishlist/abc", ``, http.StatusBadRequest},
		{"Not found", http.MethodDelete, "/wishlist/404", ``, http.StatusNotFound},
		{"Removed", http.MethodDelete, "/wishlist/1", ``, http.StatusNoContent},
		{"Short lookup", http.MethodGet, "/wishlist/lookup?kind=book&q=a", ``, http.StatusBadRequest},
		{"Catalog down", http.MethodGet, "/wishlist/lookup?kind=book&q=dune", ``, http.StatusBadGateway},
		{"Works without author", http.MethodGet, "/wishlist/works", ``, http.StatusBadRequest},
		{"Works bad sort", http.MethodGet, "/wishlist/works?authorKey=OL1A&sort=rating", ``, http.StatusBadRequest},
		{"Works bad page", http.MethodGet, "/wishlist/works?authorKey=OL1A&page=x", ``, http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			router := newRouter(&mocks.MockItemRepository{
				AddFn: func(ctx context.Context, item *models.Item) error { return models.ErrDuplicate },
				DeleteFn: func(ctx context.Context, id int64, userId int) error {
					if id == 404 {
						return models.ErrItemNotFound
					}
					return nil
				},
			}, &mocks.MockCatalogLookup{
				BooksFn: func(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
					return nil, errors.New("down")
				},
			})
			rec := httptest.NewRecorder()

			// Act
			router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.url, strings.NewReader(tc.body)))

			// Assert
			assert.Equal(t, tc.status, rec.Code)
		})
	}
}

func TestWishlistHandler_HandleLookup_Authors_ReturnsSuggestions(t *testing.T) {
	t.Parallel()

	// Arrange
	router := newRouter(&mocks.MockItemRepository{}, &mocks.MockCatalogLookup{
		AuthorsFn: func(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
			return []models.Suggestion{{Kind: models.ItemKindAuthor, Title: "Carmen Mola", AuthorKey: "OL9A"}}, nil
		},
	})
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/wishlist/lookup?kind=author&q=mola", nil))

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"authorKey":"OL9A"`)
}

func TestWishlistHandler_HandleAuthorWorks_Query_ReturnsCamelCasePage(t *testing.T) {
	t.Parallel()

	// Arrange
	var got models.AuthorWorksQuery
	router := newRouter(&mocks.MockItemRepository{}, &mocks.MockCatalogLookup{
		WorksFn: func(ctx context.Context, query models.AuthorWorksQuery, limit int) (models.WorksPage, error) {
			got = query
			return models.WorksPage{
				Works: []models.Suggestion{{Kind: models.ItemKindBook, Title: "El imperio final", WorkKey: "OL2W"}},
				Total: 44, Page: query.Page, HasMore: true,
			}, nil
		},
	})
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/wishlist/works?authorKey=OL1A&sort=new&lang=es&page=2", nil))

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, models.AuthorWorksQuery{AuthorKey: "OL1A", Sort: models.WorksSortNew, Language: "es", Page: 2}, got)
	var body struct {
		Works   []map[string]any `json:"works"`
		Total   int              `json:"total"`
		Page    int              `json:"page"`
		HasMore bool             `json:"hasMore"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Works, 1)
	assert.Equal(t, "OL2W", body.Works[0]["workKey"])
	assert.Equal(t, 44, body.Total)
	assert.Equal(t, 2, body.Page)
	assert.True(t, body.HasMore)
}
