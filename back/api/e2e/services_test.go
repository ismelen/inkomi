//go:build e2e

package e2e_test

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPreview_Page_RendersOriginalAndProcessed(t *testing.T) {
	t.Parallel()

	// Arrange
	e := loadEnv(t)

	// Act
	resp, body := get(t, e.previewURL+"/preview?sample=color&submitted=1&grayscale=1&dithering=1")

	// Assert
	require.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
	assert.Contains(t, string(body), `src="/preview/original?sample=color"`)
	assert.Contains(t, string(body), "/preview/processed?")
}

// The samples come from the production bucket, so this also checks the pipeline uploaded them.
func TestPreview_Images_ReturnDecodableImages(t *testing.T) {
	t.Parallel()

	e := loadEnv(t)
	cases := []struct {
		name, path, format string
	}{
		{"original bw", "/preview/original?sample=bw", "jpeg"},
		{"original color", "/preview/original?sample=color", "jpeg"},
		{"processed bw", "/preview/processed?sample=bw&submitted=1&grayscale=1&dithering=1&palette8bit=1", "png"},
		{"processed color", "/preview/processed?sample=color&submitted=1&sharpenEnabled=1&cropTolerance=0.02", "jpeg"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Act
			resp, body := get(t, e.previewURL+c.path)

			// Assert
			require.Equal(t, http.StatusOK, resp.StatusCode, "body: %s", body)
			cfg, format, err := image.DecodeConfig(bytes.NewReader(body))
			require.NoError(t, err)
			assert.Equal(t, c.format, format)
			assert.Positive(t, cfg.Width)
		})
	}
}

func TestPreview_InvalidConfig_Returns400(t *testing.T) {
	t.Parallel()

	// Arrange
	e := loadEnv(t)

	// Act
	resp, _ := get(t, e.previewURL+"/preview/processed?sample=bw&gamma=9")

	// Assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestProtectedEndpoints_WithoutToken_Return401(t *testing.T) {
	t.Parallel()

	e := loadEnv(t)
	cases := []struct {
		name, method, url string
	}{
		{"profile", http.MethodGet, e.authURL + "/users/me"},
		{"library search", http.MethodGet, e.libraryURL + "/books/search?q=one+piece"},
		{"upload request", http.MethodPost, e.tmURL + "/uploads/request"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Act
			status := doJSON(t, c.method, c.url, "", nil, nil)

			// Assert
			assert.Equal(t, http.StatusUnauthorized, status)
		})
	}
}

func TestAuth_RegisteredUser_CanReadProfile(t *testing.T) {
	t.Parallel()

	// Arrange
	e := loadEnv(t)
	token := registerUser(t, e)

	// Act
	var profile struct {
		Id    int    `json:"id"`
		Email string `json:"email"`
	}
	status := doJSON(t, http.MethodGet, e.authURL+"/users/me", token, nil, &profile)

	// Assert
	require.Equal(t, http.StatusOK, status)
	assert.Positive(t, profile.Id)
	assert.True(t, strings.HasSuffix(profile.Email, "@inkomi.test"))
}
