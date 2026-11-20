//go:build e2e

package e2e_test

import (
	"net/http"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRelease_Latest_PointsToTheAPK checks the update flow of the app against the real GitHub
// Releases of the repository: the latest version and a redirect to its apk.
func TestRelease_Latest_PointsToTheAPK(t *testing.T) {
	t.Parallel()

	// Arrange
	e := loadEnv(t)
	noRedirect := &http.Client{
		Timeout:       httpClient.Timeout,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}

	// Act
	var latest struct {
		Version      string `json:"version"`
		FileName     string `json:"fileName"`
		Size         int64  `json:"size"`
		DownloadPath string `json:"downloadPath"`
	}
	status := doJSON(t, http.MethodGet, e.releaseURL+"/app/latest", "", nil, &latest)
	require.Equal(t, http.StatusOK, status)

	resp, err := noRedirect.Get(e.releaseURL + latest.DownloadPath)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Assert
	assert.Regexp(t, regexp.MustCompile(`^\d+\.\d+\.\d+$`), latest.Version)
	assert.Equal(t, "inkomi.apk", latest.FileName)
	assert.Positive(t, latest.Size)
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	assert.True(t, strings.HasPrefix(location, "https://github.com/"), location)
	assert.True(t, strings.HasSuffix(location, "/v"+latest.Version+"/inkomi.apk"), location)
}
