//go:build e2e

package e2e_test

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/common/cloud"
)

// env holds the in-cluster addresses of the services under test. The Job sets them; the R2
// credentials come from the manga-worker .env mounted at /app/.env.
type env struct {
	authURL    string
	libraryURL string
	tmURL      string
	previewURL string
	releaseURL string
}

func loadEnv(t *testing.T) env {
	t.Helper()

	// .env is optional; the environment may already be set
	_ = godotenv.Load()
	return env{
		authURL:    getenv("AUTH_URL", "http://auth-service:3100"),
		libraryURL: getenv("LIBRARY_URL", "http://library-service:3001"),
		tmURL:      getenv("TM_URL", "http://transaction-manager:3000"),
		previewURL: getenv("PREVIEW_URL", "http://manga-worker:3003"),
		releaseURL: getenv("RELEASE_URL", "http://release-service:3200"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func newCloudStorage(t *testing.T) cloud.ICloudStorage {
	t.Helper()

	storage, err := cloud.NewCloudflareR2Storage(
		context.Background(),
		os.Getenv("R2_ACCOUNT_ID"),
		os.Getenv("R2_ACCESS_KEY_ID"),
		os.Getenv("R2_SECRET_ACCESS_KEY"),
		os.Getenv("R2_BUCKET_NAME"),
	)
	require.NoError(t, err)
	return storage
}

var httpClient = &http.Client{Timeout: 60 * time.Second}

// doJSON sends body as JSON with an optional bearer token and decodes the response into out.
func doJSON(t *testing.T, method, url, token string, body, out any) int {
	t.Helper()

	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	if out != nil && resp.StatusCode < 300 && len(data) > 0 {
		require.NoError(t, json.Unmarshal(data, out), "body: %s", data)
	}
	return resp.StatusCode
}

// registerUser creates a throwaway user and returns its access token. The sandbox database is
// dropped with its namespace, so nothing has to be cleaned up.
func registerUser(t *testing.T, e env) string {
	t.Helper()

	var tokens struct {
		AccessToken string `json:"accessToken"`
	}
	status := doJSON(t, http.MethodPost, e.authURL+"/auth/register", "", map[string]string{
		"email":    fmt.Sprintf("e2e-%d@inkomi.test", time.Now().UnixNano()),
		"password": "e2e-password-123",
		"deviceId": "e2e",
	}, &tokens)
	require.Less(t, status, 300, "register status")
	require.NotEmpty(t, tokens.AccessToken)
	return tokens.AccessToken
}

// buildCBZ returns a small manga chapter: two portrait pages and a double page spread.
func buildCBZ(t *testing.T) []byte {
	t.Helper()

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	pages := []struct {
		name string
		w, h int
	}{
		{"001.jpg", 800, 1200},
		{"002.jpg", 1600, 1200},
		{"003.jpg", 800, 1200},
	}
	for i, p := range pages {
		w, err := zw.Create(p.name)
		require.NoError(t, err)
		require.NoError(t, jpeg.Encode(w, testPage(p.w, p.h, i), &jpeg.Options{Quality: 85}))
	}
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

func testPage(w, h, seed int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: color.White}, image.Point{}, draw.Src)
	ink := image.Rect(w/10, h/10, w-w/10, h-h/10)
	draw.Draw(img, ink, &image.Uniform{C: color.Gray{Y: uint8(40 * (seed + 1))}}, image.Point{}, draw.Src)
	return img
}

func putFile(t *testing.T, url string, data []byte) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
	require.NoError(t, err)
	req.ContentLength = int64(len(data))

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	require.Less(t, resp.StatusCode, 300, "upload status, body: %s", body)
}

func get(t *testing.T, url string) (*http.Response, []byte) {
	t.Helper()

	resp, err := httpClient.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp, body
}
