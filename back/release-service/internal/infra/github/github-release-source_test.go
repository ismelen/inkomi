package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const releaseJSON = `{
	"tag_name": "v1.3.7",
	"body": "notes",
	"published_at": "2026-10-01T10:00:00Z",
	"assets": [
		{"name": "other.zip", "browser_download_url": "https://x/other.zip", "size": 1},
		{"name": "inkomi.apk", "browser_download_url": "https://x/inkomi.apk", "size": 4096, "digest": "sha256:abc"}
	]
}`

func server(t *testing.T, status int, body string, gotAuth *string) *httptest.Server {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/owner/repo/releases/latest" {
			http.NotFound(w, r)
			return
		}
		if gotAuth != nil {
			*gotAuth = r.Header.Get("Authorization")
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestGitHubReleaseSource_Latest_WithAPK_ReturnsRelease(t *testing.T) {
	t.Parallel()

	// Arrange
	var auth string
	srv := server(t, http.StatusOK, releaseJSON, &auth)
	source := github.NewGitHubReleaseSource(srv.URL, "owner/repo", "secret", "inkomi.apk")

	// Act
	release, err := source.Latest(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "Bearer secret", auth)
	assert.Equal(t, "1.3.7", release.Version.String())
	assert.Equal(t, "notes", release.Notes)
	assert.Equal(t, 2026, release.PublishedAt.Year())
	assert.Equal(t, "https://x/inkomi.apk", release.APK.URL)
	assert.Equal(t, int64(4096), release.APK.Size)
	assert.Equal(t, "abc", release.APK.SHA256)
}

func TestGitHubReleaseSource_Latest_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		body   string
		asset  string
		want   error
	}{
		{"no releases", http.StatusNotFound, `{}`, "inkomi.apk", models.ErrReleaseNotFound},
		{"rate limited", http.StatusForbidden, `{"message":"rate limit"}`, "inkomi.apk", models.ErrSourceUnavailable},
		{"missing asset", http.StatusOK, releaseJSON, "missing.apk", models.ErrAPKNotFound},
		{"bad tag", http.StatusOK, `{"tag_name":"latest","assets":[]}`, "inkomi.apk", models.ErrInvalidVersion},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			srv := server(t, tt.status, tt.body, nil)
			source := github.NewGitHubReleaseSource(srv.URL, "owner/repo", "", tt.asset)

			// Act
			release, err := source.Latest(context.Background())

			// Assert
			assert.ErrorIs(t, err, tt.want)
			assert.Nil(t, release)
		})
	}
}

func TestGitHubReleaseSource_Latest_Unreachable_ReturnsSourceUnavailable(t *testing.T) {
	t.Parallel()

	// Arrange
	source := github.NewGitHubReleaseSource("http://127.0.0.1:1", "owner/repo", "", "inkomi.apk")

	// Act
	_, err := source.Latest(context.Background())

	// Assert
	assert.ErrorIs(t, err, models.ErrSourceUnavailable)
}
