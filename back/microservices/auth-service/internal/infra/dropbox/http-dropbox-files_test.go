package dropbox_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/microservices/auth-service/internal/infra/dropbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPDropboxFiles_ListFolders_Pages_ReturnsSortedFoldersOnly(t *testing.T) {
	t.Parallel()

	// Arrange
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer tok", r.Header.Get("Authorization"))
		var body map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		switch r.URL.Path {
		case "/2/files/list_folder":
			assert.Equal(t, "/Books", body["path"])
			_, _ = w.Write([]byte(`{"entries":[{".tag":"folder","name":"manga","path_display":"/Books/manga"},{".tag":"file","name":"a.epub","path_display":"/Books/a.epub"}],"cursor":"c1","has_more":true}`))
		case "/2/files/list_folder/continue":
			assert.Equal(t, "c1", body["cursor"])
			_, _ = w.Write([]byte(`{"entries":[{".tag":"folder","name":"Kobo","path_display":"/Books/Kobo"}],"has_more":false}`))
		default:
			t.Errorf("unexpected route %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	// Act
	folders, err := dropbox.NewHTTPDropboxFiles(srv.Client(), srv.URL).ListFolders(context.Background(), "tok", "/Books")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []models.DropboxFolder{{Name: "Kobo", Path: "/Books/Kobo"}, {Name: "manga", Path: "/Books/manga"}}, folders)
}

func TestHTTPDropboxFiles_Errors_MapToDomain(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"Missing path", http.StatusConflict, `{"error_summary":"path/not_found/.."}`, models.ErrDropboxFolderMissing},
		{"Existing folder", http.StatusConflict, `{"error_summary":"path/conflict/folder/.."}`, models.ErrDropboxFolderExists},
		{"Revoked token", http.StatusUnauthorized, `{"error_summary":"expired_access_token/"}`, models.ErrInvalidExternalToken},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(tc.body))
			}))
			defer srv.Close()

			// Act
			_, err := dropbox.NewHTTPDropboxFiles(srv.Client(), srv.URL).CreateFolder(context.Background(), "tok", "/a")

			// Assert
			assert.ErrorIs(t, err, tc.want)
		})
	}
}
