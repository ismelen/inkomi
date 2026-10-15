package dropbox_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/push-service/internal/infra/dropbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type call struct {
	endpoint string
	arg      map[string]any
	rawArg   string
	auth     string
	body     string
}

type fakeDropbox struct {
	mu     sync.Mutex
	calls  []call
	status int
	body   string
}

func (f *fakeDropbox) start(t *testing.T) string {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)

		rawArg := r.Header.Get("Dropbox-API-Arg")
		var arg map[string]any
		require.NoError(t, json.Unmarshal([]byte(rawArg), &arg))

		f.mu.Lock()
		f.calls = append(f.calls, call{r.URL.Path, arg, rawArg, r.Header.Get("Authorization"), string(body)})
		f.mu.Unlock()

		if f.status != 0 {
			w.WriteHeader(f.status)
			_, _ = w.Write([]byte(f.body))
			return
		}
		if r.URL.Path == "/files/upload_session/start" {
			_, _ = w.Write([]byte(`{"session_id":"s1"}`))
			return
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func writeFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func account(t *testing.T) *models.DropboxAccount {
	t.Helper()

	account, err := models.NewDropboxAccount("dbx", "/Inkomi")
	require.NoError(t, err)
	return account
}

func TestDropboxCloudDrive_UploadFile_SmallFile_SingleUpload(t *testing.T) {
	t.Parallel()

	// Arrange
	fake := &fakeDropbox{}
	drive := dropbox.NewDropboxCloudDrive(http.DefaultClient, fake.start(t), 16)

	// Act
	err := drive.UploadFile(context.Background(), account(t), writeFile(t, "epub"), "Kōkaku – 攻殻.epub")

	// Assert
	require.NoError(t, err)
	require.Len(t, fake.calls, 1)
	c := fake.calls[0]
	assert.Equal(t, "/files/upload", c.endpoint)
	assert.Equal(t, "Bearer dbx", c.auth)
	assert.Equal(t, "epub", c.body)
	assert.Equal(t, "/Inkomi/Kōkaku – 攻殻.epub", c.arg["path"])
	assert.Equal(t, "overwrite", c.arg["mode"])
	assert.NotContains(t, c.rawArg, "攻")
}

func TestDropboxCloudDrive_UploadFile_BigFile_UploadSession(t *testing.T) {
	t.Parallel()

	// Arrange
	fake := &fakeDropbox{}
	drive := dropbox.NewDropboxCloudDrive(http.DefaultClient, fake.start(t), 4)

	// Act
	err := drive.UploadFile(context.Background(), account(t), writeFile(t, "0123456789"), "book.epub")

	// Assert
	require.NoError(t, err)
	require.Len(t, fake.calls, 3)

	assert.Equal(t, "/files/upload_session/start", fake.calls[0].endpoint)
	assert.Equal(t, "0123", fake.calls[0].body)

	assert.Equal(t, "/files/upload_session/append_v2", fake.calls[1].endpoint)
	assert.Equal(t, "4567", fake.calls[1].body)
	assert.Equal(t, map[string]any{"session_id": "s1", "offset": float64(4)}, fake.calls[1].arg["cursor"])

	assert.Equal(t, "/files/upload_session/finish", fake.calls[2].endpoint)
	assert.Equal(t, "89", fake.calls[2].body)
	assert.Equal(t, map[string]any{"session_id": "s1", "offset": float64(8)}, fake.calls[2].arg["cursor"])
	commit := fake.calls[2].arg["commit"].(map[string]any)
	assert.Equal(t, "/Inkomi/book.epub", commit["path"])
	assert.Equal(t, "overwrite", commit["mode"])
}

func TestDropboxCloudDrive_UploadFile_ErrorStatus_Rejected(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		status   int
		body     string
		rejected bool
	}{
		{"Conflict", http.StatusConflict, `{"error_summary":"path/insufficient_space/"}`, true},
		{"Unauthorized", http.StatusUnauthorized, `{"error_summary":"expired_access_token/"}`, true},
		{"Bad request", http.StatusBadRequest, "Error in call to API function", true},
		{"Rate limited", http.StatusTooManyRequests, `{"error_summary":"too_many_requests/"}`, false},
		{"Server error", http.StatusServiceUnavailable, "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			fake := &fakeDropbox{status: tc.status, body: tc.body}
			drive := dropbox.NewDropboxCloudDrive(http.DefaultClient, fake.start(t), 16)

			// Act
			err := drive.UploadFile(context.Background(), account(t), writeFile(t, "epub"), "book.epub")

			// Assert
			require.Error(t, err)
			assert.Equal(t, tc.rejected, errors.Is(err, models.ErrUploadRejected))
		})
	}
}
