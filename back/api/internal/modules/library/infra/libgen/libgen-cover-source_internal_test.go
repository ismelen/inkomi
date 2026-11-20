package libgen

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hotlinkServer answers like a libgen mirror: an empty image unless the Referer is its own site.
func hotlinkServer(t *testing.T) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		if r.Header.Get("Referer") != server.URL+"/" {
			w.Header().Set("Content-Length", "0")
			return
		}
		_, _ = w.Write([]byte("jpeg"))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestFetchCover_HotlinkProtectedMirror_SendsRefererAndReadsImage(t *testing.T) {
	t.Parallel()

	// Arrange
	server := hotlinkServer(t)

	// Act
	cover, err := fetchCover(context.Background(), server.URL+"/covers/1/abc_small.jpg")

	// Assert
	require.NoError(t, err)
	defer cover.Body.Close()
	body, err := io.ReadAll(cover.Body)
	require.NoError(t, err)
	assert.Equal(t, "jpeg", string(body))
	assert.Equal(t, "image/jpeg", cover.ContentType)
}

func TestFetchCover_EmptyImage_ReturnsErrInvalidCoverURL(t *testing.T) {
	t.Parallel()

	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Content-Length", "0")
	}))
	t.Cleanup(server.Close)

	// Act
	cover, err := fetchCover(context.Background(), server.URL+"/covers/1/abc_small.jpg")

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidCoverURL)
	assert.Nil(t, cover)
}
