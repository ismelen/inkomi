package dropbox_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/dropbox"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newServer(t *testing.T, status int, body string, check func(r *http.Request)) *dropbox.HTTPDropboxOAuth {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		check(r)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	return dropbox.NewHTTPDropboxOAuth(server.Client(), server.URL, "app-key")
}

func TestHTTPDropboxOAuth_Refresh_Ok_ReturnsAccessToken(t *testing.T) {
	t.Parallel()

	// Arrange
	oauth := newServer(t, http.StatusOK, `{"access_token":"sl.access","expires_in":14400,"token_type":"bearer"}`, func(r *http.Request) {
		require.NoError(t, r.ParseForm())
		assert.Equal(t, "/oauth2/token", r.URL.Path)
		assert.Equal(t, "refresh_token", r.PostForm.Get("grant_type"))
		assert.Equal(t, "dbx-refresh", r.PostForm.Get("refresh_token"))
		assert.Equal(t, "app-key", r.PostForm.Get("client_id"))
	})

	// Act
	access, err := oauth.Refresh(context.Background(), "dbx-refresh")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, "sl.access", access.Token)
	assert.WithinDuration(t, time.Now().Add(4*time.Hour), access.ExpiresAt, time.Minute)
}

func TestHTTPDropboxOAuth_Account_Ok_ReturnsExternalAccount(t *testing.T) {
	t.Parallel()

	// Arrange
	oauth := newServer(t, http.StatusOK, `{"account_id":"dbid:1","email":"a@b.com","email_verified":true}`, func(r *http.Request) {
		assert.Equal(t, "/2/users/get_current_account", r.URL.Path)
		assert.Equal(t, "Bearer sl.access", r.Header.Get("Authorization"))
	})

	// Act
	account, err := oauth.Account(context.Background(), "sl.access")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, &models.ExternalAccount{Provider: models.ProviderDropbox, Subject: "dbid:1", Email: "a@b.com", EmailVerified: true}, account)
}

func TestHTTPDropboxOAuth_ErrorStatus_InvalidExternalToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		status  int
		invalid bool
	}{
		{"Bad request", http.StatusBadRequest, true},
		{"Unauthorized", http.StatusUnauthorized, true},
		{"Server error", http.StatusInternalServerError, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			oauth := newServer(t, tc.status, `{"error":"invalid_grant"}`, func(r *http.Request) {})

			// Act
			_, refreshErr := oauth.Refresh(context.Background(), "dbx-refresh")
			_, accountErr := oauth.Account(context.Background(), "sl.access")

			// Assert
			require.Error(t, refreshErr)
			require.Error(t, accountErr)
			assert.Equal(t, tc.invalid, errors.Is(refreshErr, models.ErrInvalidExternalToken))
			assert.Equal(t, tc.invalid, errors.Is(accountErr, models.ErrInvalidExternalToken))
		})
	}
}

func TestHTTPDropboxOAuth_Refresh_EmptyToken_InvalidExternalToken(t *testing.T) {
	t.Parallel()

	// Arrange
	oauth := dropbox.NewHTTPDropboxOAuth(http.DefaultClient, "http://unused", "app-key")

	// Act
	_, err := oauth.Refresh(context.Background(), "")

	// Assert
	assert.ErrorIs(t, err, models.ErrInvalidExternalToken)
}
