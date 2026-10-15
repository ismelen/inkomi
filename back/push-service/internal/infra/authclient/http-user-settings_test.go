package authclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ismelen/inkomi/back/push-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/push-service/internal/infra/authclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSigner struct{}

func (fakeSigner) Sign(audience string) (string, error) { return "signed-for-" + audience, nil }

type received struct {
	path string
	auth string
}

func newSettings(t *testing.T, status int, body string) (*authclient.HTTPUserSettings, *received) {
	t.Helper()

	var req received
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req = received{path: r.URL.Path, auth: r.Header.Get("Authorization")}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	return authclient.NewHTTPUserSettings(server.Client(), server.URL+"/", "auth-service", fakeSigner{}), &req
}

func TestHTTPUserSettings_GetDropboxAccount_Ok_ReturnsAccount(t *testing.T) {
	t.Parallel()

	// Arrange
	settings, req := newSettings(t, http.StatusOK, `{"token":"dbx","folder":"Inkomi/"}`)

	// Act
	account, err := settings.GetDropboxAccount(context.Background(), 7)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, &models.DropboxAccount{Token: "dbx", Folder: "/Inkomi"}, account)
	assert.Equal(t, "/internal/users/7/dropbox", req.path)
	assert.Equal(t, "Bearer signed-for-auth-service", req.auth)
}

func TestHTTPUserSettings_GetFCMToken_Ok_ReturnsToken(t *testing.T) {
	t.Parallel()

	// Arrange
	settings, req := newSettings(t, http.StatusOK, `{"token":"fcm"}`)

	// Act
	token, err := settings.GetFCMToken(context.Background(), 7)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "fcm", token)
	assert.Equal(t, "/internal/users/7/fcm-token", req.path)
}

func TestHTTPUserSettings_Responses_Errors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		status     int
		body       string
		dropboxErr error
		fcmErr     error
	}{
		{"Not found", http.StatusNotFound, "", models.ErrDropboxNotLinked, models.ErrFCMTokenNotFound},
		{"Empty token", http.StatusOK, `{"token":""}`, models.ErrDropboxNotLinked, models.ErrFCMTokenNotFound},
		{"Unauthorized", http.StatusUnauthorized, "invalid token", nil, nil},
		{"Server error", http.StatusInternalServerError, "boom", nil, nil},
		{"Invalid json", http.StatusOK, "{", nil, nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			settings, _ := newSettings(t, tc.status, tc.body)

			// Act
			_, dropboxErr := settings.GetDropboxAccount(context.Background(), 7)
			_, fcmErr := settings.GetFCMToken(context.Background(), 7)

			// Assert
			require.Error(t, dropboxErr)
			require.Error(t, fcmErr)
			if tc.dropboxErr != nil {
				assert.ErrorIs(t, dropboxErr, tc.dropboxErr)
				assert.ErrorIs(t, fcmErr, tc.fcmErr)
			} else {
				assert.NotErrorIs(t, dropboxErr, models.ErrDropboxNotLinked)
				assert.NotErrorIs(t, fcmErr, models.ErrFCMTokenNotFound)
			}
		})
	}
}
