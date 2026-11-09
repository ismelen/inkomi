package routes_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/api/handlers"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/api/routes"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/infra/security"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/dropboxlogin"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/getprofile"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/googlelogin"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/linkdropbox"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/login"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/logout"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/mocks"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/refreshsession"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/register"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/socialsignin"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/startsession"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/unlinkdropbox"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/updatedropboxfolder"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/auth/usecases/updatefcmtoken"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/httpauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type keyPair struct {
	privatePEM string
	publicPEM  string
}

func newKeyPair(t *testing.T) keyPair {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	require.NoError(t, err)
	return keyPair{
		privatePEM: string(pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})),
		publicPEM:  string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})),
	}
}

type server struct {
	api          http.Handler
	userToken    string
	foreignToken string
}

// newServer wires the real handlers and use cases over in-memory mocks with user 7, who has
// an FCM token and no Dropbox account.
func newServer(t *testing.T) *server {
	t.Helper()

	userKeys, foreignKeys := newKeyPair(t), newKeyPair(t)
	issuer, err := security.NewJWTTokenIssuer("auth-service", userKeys.privatePEM, time.Minute)
	require.NoError(t, err)
	foreignIssuer, err := security.NewJWTTokenIssuer("auth-service", foreignKeys.privatePEM, time.Minute)
	require.NoError(t, err)

	user := &models.User{Id: 7, Email: "a@b.com", PasswordHash: "hash:secret-password", Role: models.RoleUser, FCMToken: "fcm"}
	users := &mocks.MockUserRepository{
		GetByIdFn: func(ctx context.Context, id int) (*models.User, error) {
			if id != user.Id {
				return nil, models.ErrUserNotFound
			}
			return user, nil
		},
		GetByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			if email != user.Email {
				return nil, models.ErrUserNotFound
			}
			return user, nil
		},
		CreateFn: func(ctx context.Context, u *models.User) error {
			if u.Email == user.Email {
				return models.ErrEmailTaken
			}
			u.Id = 8
			return nil
		},
	}
	sessions, identities := &mocks.MockSessionRepository{}, &mocks.MockIdentityRepository{}
	links := &mocks.MockDropboxLinkRepository{
		UpdateFolderFn: func(ctx context.Context, userId int, folder string) error { return models.ErrDropboxNotLinked },
	}
	hasher, oauth := &mocks.MockPasswordHasher{}, &mocks.MockDropboxOAuth{}

	startSession := startsession.NewStartSessionUC(sessions, issuer, time.Hour)
	signIn := socialsignin.NewSocialSignInUC(&mocks.MockTxManager{}, users, identities)
	authHandler := handlers.NewAuthHandler(
		register.NewRegisterUC(users, hasher, startSession),
		login.NewLoginUC(users, hasher, startSession),
		googlelogin.NewGoogleLoginUC(&mocks.MockGoogleVerifier{}, signIn, startSession),
		dropboxlogin.NewDropboxLoginUC(oauth, links, signIn, startSession),
		refreshsession.NewRefreshSessionUC(sessions, users, issuer, time.Hour),
		logout.NewLogoutUC(sessions, issuer),
	)
	usersHandler := handlers.NewUsersHandler(
		getprofile.NewGetProfileUC(users, identities, links),
		updatefcmtoken.NewUpdateFCMTokenUC(users),
		linkdropbox.NewLinkDropboxUC(oauth, identities, links),
		updatedropboxfolder.NewUpdateDropboxFolderUC(links),
		unlinkdropbox.NewUnlinkDropboxUC(links),
	)

	authMiddleware, err := httpauth.NewAuthMiddleware(userKeys.publicPEM)
	require.NoError(t, err)

	api := chi.NewRouter()
	routes.SetupAuthRoutes(api, authHandler)
	routes.SetupUserRoutes(api, usersHandler, authMiddleware)

	access, err := issuer.IssueAccess(user)
	require.NoError(t, err)
	foreign, err := foreignIssuer.IssueAccess(user)
	require.NoError(t, err)

	return &server{api: api, userToken: access.Token, foreignToken: foreign.Token}
}

func do(handler http.Handler, method, path, token, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func TestRoutes_Requests_Status(t *testing.T) {
	t.Parallel()

	s := newServer(t)
	cases := []struct {
		name   string
		method string
		path   string
		token  func(s *server) string
		body   string
		status int
	}{
		{"Register", http.MethodPost, "/auth/register", nil, `{"email":"new@b.com","password":"secret-password","deviceId":"d"}`, http.StatusOK},
		{"Register taken email", http.MethodPost, "/auth/register", nil, `{"email":"a@b.com","password":"secret-password","deviceId":"d"}`, http.StatusConflict},
		{"Register weak password", http.MethodPost, "/auth/register", nil, `{"email":"new@b.com","password":"x","deviceId":"d"}`, http.StatusBadRequest},
		{"Register invalid json", http.MethodPost, "/auth/register", nil, `{`, http.StatusBadRequest},
		{"Login", http.MethodPost, "/auth/login", nil, `{"email":"a@b.com","password":"secret-password","deviceId":"d"}`, http.StatusOK},
		{"Login wrong password", http.MethodPost, "/auth/login", nil, `{"email":"a@b.com","password":"wrong-password","deviceId":"d"}`, http.StatusUnauthorized},
		{"Google invalid token", http.MethodPost, "/auth/google", nil, `{"idToken":"x","deviceId":"d"}`, http.StatusUnauthorized},
		{"Dropbox invalid token", http.MethodPost, "/auth/dropbox", nil, `{"refreshToken":"x","deviceId":"d"}`, http.StatusUnauthorized},
		{"Refresh unknown token", http.MethodPost, "/auth/refresh", nil, `{"refreshToken":"x"}`, http.StatusUnauthorized},
		{"Logout", http.MethodPost, "/auth/logout", nil, `{"refreshToken":"x"}`, http.StatusNoContent},
		{"Profile without token", http.MethodGet, "/users/me", nil, "", http.StatusUnauthorized},
		{"Profile with foreign token", http.MethodGet, "/users/me", func(s *server) string { return s.foreignToken }, "", http.StatusUnauthorized},
		{"Profile", http.MethodGet, "/users/me", func(s *server) string { return s.userToken }, "", http.StatusOK},
		{"Update fcm token", http.MethodPut, "/users/me/fcm-token", func(s *server) string { return s.userToken }, `{"token":"new"}`, http.StatusNoContent},
		{"Folder invalid", http.MethodPut, "/users/me/dropbox/folder", func(s *server) string { return s.userToken }, `{"folder":"/a/../b"}`, http.StatusBadRequest},
		{"Folder not linked", http.MethodPut, "/users/me/dropbox/folder", func(s *server) string { return s.userToken }, `{"folder":"/a"}`, http.StatusNotFound},
		{"Unlink dropbox", http.MethodDelete, "/users/me/dropbox", func(s *server) string { return s.userToken }, "", http.StatusNoContent},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			token := ""
			if tc.token != nil {
				token = tc.token(s)
			}

			// Act
			rec := do(s.api, tc.method, tc.path, token, tc.body)

			// Assert
			assert.Equal(t, tc.status, rec.Code, rec.Body.String())
		})
	}
}

func TestRoutes_Login_TokensResponseInCamelCase(t *testing.T) {
	t.Parallel()

	// Arrange
	s := newServer(t)

	// Act
	rec := do(s.api, http.MethodPost, "/auth/login", "", `{"email":"a@b.com","password":"secret-password","deviceId":"d"}`)

	// Assert
	require.Equal(t, http.StatusOK, rec.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.NotEmpty(t, body["accessToken"])
	assert.NotEmpty(t, body["refreshToken"])
	assert.NotEmpty(t, body["accessTokenExpiresAt"])
	assert.NotEmpty(t, body["refreshTokenExpiresAt"])
	assert.Equal(t, "no-store", rec.Header().Get("Cache-Control"))
}
