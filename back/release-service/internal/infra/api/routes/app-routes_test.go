package routes_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/release-service/internal/usecases/getlatestrelease"
	"github.com/ismelen/inkomi/back/release-service/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func router(release *models.Release, err error) http.Handler {
	source := &mocks.MockReleaseSource{
		LatestFn: func(ctx context.Context) (*models.Release, error) { return release, err },
	}
	api := chi.NewRouter()
	routes.SetupAppRoutes(api, handlers.NewAppHandler(getlatestrelease.NewGetLatestReleaseUC(source)))
	return api
}

func do(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

var latest = &models.Release{
	Version: models.Version{Major: 2, Minor: 1},
	Notes:   "notes",
	APK:     models.ReleaseAsset{Name: "inkomi.apk", URL: "https://github.com/x/inkomi.apk", Size: 10, SHA256: "abc"},
}

func TestAppRoutes_Latest_ReturnsCamelCaseRelease(t *testing.T) {
	t.Parallel()

	// Act
	rec := do(router(latest, nil), "/app/latest")

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "2.1.0", body["version"])
	assert.Equal(t, "inkomi.apk", body["fileName"])
	assert.Equal(t, "abc", body["sha256"])
	assert.Equal(t, "/app/download", body["downloadPath"])
	assert.NotContains(t, body, "url")
}

func TestAppRoutes_Download_RedirectsToAPK(t *testing.T) {
	t.Parallel()

	// Act
	rec := do(router(latest, nil), "/app/download")

	// Assert
	assert.Equal(t, http.StatusFound, rec.Code)
	assert.Equal(t, "https://github.com/x/inkomi.apk", rec.Header().Get("Location"))
}

func TestAppRoutes_Errors_MapToStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
	}{
		{"no release", models.ErrReleaseNotFound, http.StatusNotFound},
		{"no apk", models.ErrAPKNotFound, http.StatusNotFound},
		{"github down", fmt.Errorf("%w: boom", models.ErrSourceUnavailable), http.StatusServiceUnavailable},
		{"unexpected", models.ErrInvalidVersion, http.StatusInternalServerError},
	}

	for _, tt := range tests {
		for _, path := range []string{"/app/latest", "/app/download"} {
			t.Run(tt.name+" "+path, func(t *testing.T) {
				t.Parallel()

				// Act
				rec := do(router(nil, tt.err), path)

				// Assert
				assert.Equal(t, tt.status, rec.Code)
				assert.Contains(t, rec.Body.String(), `"error"`)
			})
		}
	}
}
