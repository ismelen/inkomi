package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/dismissrecommendation"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/getrecommendations"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/platform/httpauth"
)

func newRouter(recs *mocks.MockRecommendationRepository, signals *mocks.MockSignalRepository) http.Handler {
	router := chi.NewRouter()
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(httpauth.WithUserClaims(r.Context(), &httpauth.UserClaims{Id: 4})))
		})
	})
	routes.SetupRecommendationsRoutes(router, handlers.NewRecommendationsHandler(
		getrecommendations.NewGetRecommendationsUC(recs), dismissrecommendation.NewDismissRecommendationUC(signals),
	))
	return router
}

func TestRecommendationsHandler_HandleList_Section_ReturnsWorksWithReasons(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotSection models.Section
	router := newRouter(&mocks.MockRecommendationRepository{
		ListFn: func(ctx context.Context, userId int, section models.Section) ([]models.Recommendation, error) {
			gotSection = section
			assert.Equal(t, 4, userId)
			return []models.Recommendation{{
				Work:   models.Work{Key: "OL1W", Title: "Las madres", Authors: []string{"Carmen Mola"}, CoverId: 7, FirstPublishYear: 2022},
				Reason: &models.Work{Key: "OL2W", Title: "La nena"},
			}}, nil
		},
	}, &mocks.MockSignalRepository{})
	rec := httptest.NewRecorder()

	// Act
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/recommendations?section=new", nil))

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, models.SectionNew, gotSection)
	var body []map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body, 1)
	assert.Equal(t, "https://covers.openlibrary.org/b/id/7-M.jpg", body[0]["coverUrl"])
	assert.Equal(t, []any{}, body[0]["subjects"])
	assert.Equal(t, "La nena", body[0]["reason"].(map[string]any)["title"])
}

func TestRecommendationsHandler_Requests_MapsStatus(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		method string
		url    string
		status int
	}{
		{"Default section", http.MethodGet, "/recommendations", http.StatusOK},
		{"Invalid section", http.MethodGet, "/recommendations?section=popular", http.StatusBadRequest},
		{"Dismiss", http.MethodPost, "/recommendations/OL1W/dismiss", http.StatusNoContent},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			var dismissed string
			router := newRouter(&mocks.MockRecommendationRepository{}, &mocks.MockSignalRepository{
				DismissFn: func(ctx context.Context, userId int, key string) error {
					dismissed = key
					return nil
				},
			})
			rec := httptest.NewRecorder()

			// Act
			router.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.url, nil))

			// Assert
			assert.Equal(t, tc.status, rec.Code)
			if tc.method == http.MethodPost {
				assert.Equal(t, "OL1W", dismissed)
			}
		})
	}
}
