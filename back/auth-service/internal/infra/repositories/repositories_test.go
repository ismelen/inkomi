package repositories_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/datasources"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/repositories"
	"github.com/ismelen/inkomi/back/auth-service/internal/infra/security"
	"github.com/ismelen/inkomi/back/auth-service/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDatasource(t *testing.T) *datasources.SQLiteDatasource {
	t.Helper()

	ds, err := datasources.NewSQLiteDatasource(filepath.Join(t.TempDir(), "users.db"), migrations.Users)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ds.Close() })
	return ds
}

func createUser(t *testing.T, users *repositories.SQLiteUserRepository, email string) *models.User {
	t.Helper()

	user := models.NewPasswordUser(email, "hash")
	require.NoError(t, users.Create(context.Background(), user))
	return user
}

func TestSQLiteUserRepository_CreateAndGet_RoundTrip(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	users := repositories.NewSQLiteUserRepository(newDatasource(t).DB)
	social := models.NewSocialUser(models.ExternalAccount{})

	// Act
	created := createUser(t, users, "a@b.com")
	require.NoError(t, users.Create(ctx, social))
	byId, errId := users.GetById(ctx, created.Id)
	byEmail, errEmail := users.GetByEmail(ctx, "a@b.com")
	noEmail, errSocial := users.GetById(ctx, social.Id)

	// Assert
	require.NoError(t, errId)
	require.NoError(t, errEmail)
	require.NoError(t, errSocial)
	assert.Positive(t, created.Id)
	assert.Equal(t, created, byId)
	assert.Equal(t, created, byEmail)
	assert.Equal(t, "", noEmail.Email)
	assert.NotEqual(t, created.Id, social.Id)
}

func TestSQLiteUserRepository_Create_DuplicatedEmail_ReturnsErrEmailTaken(t *testing.T) {
	t.Parallel()

	// Arrange
	users := repositories.NewSQLiteUserRepository(newDatasource(t).DB)
	createUser(t, users, "a@b.com")

	// Act
	err := users.Create(context.Background(), models.NewPasswordUser("a@b.com", "hash"))

	// Assert
	assert.ErrorIs(t, err, models.ErrEmailTaken)
}

func TestSQLiteUserRepository_Updates_PersistChanges(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	users := repositories.NewSQLiteUserRepository(newDatasource(t).DB)
	user := createUser(t, users, "a@b.com")
	user.ClaimEmail()

	// Act
	require.NoError(t, users.Update(ctx, user))
	require.NoError(t, users.UpdateFCMToken(ctx, user.Id, "fcm"))
	got, err := users.GetById(ctx, user.Id)

	// Assert
	require.NoError(t, err)
	assert.True(t, got.EmailVerified)
	assert.Equal(t, "", got.PasswordHash)
	assert.Equal(t, "fcm", got.FCMToken)
}

func TestSQLiteUserRepository_UnknownUser_ReturnsErrUserNotFound(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	users := repositories.NewSQLiteUserRepository(newDatasource(t).DB)

	// Act
	_, getErr := users.GetById(ctx, 99)
	_, emailErr := users.GetByEmail(ctx, "x@y.com")
	updateErr := users.Update(ctx, &models.User{Id: 99})
	fcmErr := users.UpdateFCMToken(ctx, 99, "fcm")

	// Assert
	assert.ErrorIs(t, getErr, models.ErrUserNotFound)
	assert.ErrorIs(t, emailErr, models.ErrUserNotFound)
	assert.ErrorIs(t, updateErr, models.ErrUserNotFound)
	assert.ErrorIs(t, fcmErr, models.ErrUserNotFound)
}

func TestSQLiteIdentityRepository_CreateGetList_RoundTrip(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	ds := newDatasource(t)
	users := repositories.NewSQLiteUserRepository(ds.DB)
	identities := repositories.NewSQLiteIdentityRepository(ds.DB)
	user := createUser(t, users, "a@b.com")

	// Act
	require.NoError(t, identities.Create(ctx, models.Identity{Provider: models.ProviderGoogle, Subject: "g-1", UserId: user.Id}))
	require.NoError(t, identities.Create(ctx, models.Identity{Provider: models.ProviderDropbox, Subject: "dbid:1", UserId: user.Id}))
	duplicateErr := identities.Create(ctx, models.Identity{Provider: models.ProviderGoogle, Subject: "g-1", UserId: user.Id})
	identity, err := identities.Get(ctx, models.ProviderGoogle, "g-1")
	_, missingErr := identities.Get(ctx, models.ProviderGoogle, "g-2")
	providers, listErr := identities.ListProviders(ctx, user.Id)

	// Assert
	require.NoError(t, err)
	require.NoError(t, listErr)
	assert.Error(t, duplicateErr)
	assert.Equal(t, &models.Identity{Provider: models.ProviderGoogle, Subject: "g-1", UserId: user.Id}, identity)
	assert.ErrorIs(t, missingErr, models.ErrIdentityNotFound)
	assert.Equal(t, []models.Provider{models.ProviderDropbox, models.ProviderGoogle}, providers)
}

func TestSQLiteSessionRepository_Rotate_Session_Result(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	ds := newDatasource(t)
	users := repositories.NewSQLiteUserRepository(ds.DB)
	sessions := repositories.NewSQLiteSessionRepository(ds.DB)
	user := createUser(t, users, "a@b.com")
	now := time.Now()

	require.NoError(t, sessions.Save(ctx, models.Session{UserId: user.Id, DeviceId: "d1", RefreshHash: "replaced", ExpiresAt: now.Add(time.Hour)}))
	require.NoError(t, sessions.Save(ctx, models.Session{UserId: user.Id, DeviceId: "d1", RefreshHash: "h1", ExpiresAt: now.Add(time.Hour)}))
	require.NoError(t, sessions.Save(ctx, models.Session{UserId: user.Id, DeviceId: "d2", RefreshHash: "expired", ExpiresAt: now.Add(-time.Hour)}))

	// Act
	userId, err := sessions.Rotate(ctx, "h1", "h2", now, now.Add(time.Hour))
	_, reusedErr := sessions.Rotate(ctx, "h1", "h3", now, now.Add(time.Hour))
	_, replacedErr := sessions.Rotate(ctx, "replaced", "h4", now, now.Add(time.Hour))
	_, expiredErr := sessions.Rotate(ctx, "expired", "h5", now, now.Add(time.Hour))
	require.NoError(t, sessions.DeleteByRefreshHash(ctx, "h2"))
	_, loggedOutErr := sessions.Rotate(ctx, "h2", "h6", now, now.Add(time.Hour))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, user.Id, userId)
	assert.ErrorIs(t, reusedErr, models.ErrInvalidRefreshToken)
	assert.ErrorIs(t, replacedErr, models.ErrInvalidRefreshToken)
	assert.ErrorIs(t, expiredErr, models.ErrInvalidRefreshToken)
	assert.ErrorIs(t, loggedOutErr, models.ErrInvalidRefreshToken)
}

func newDropboxRepository(t *testing.T, ds *datasources.SQLiteDatasource) *repositories.SQLiteDropboxLinkRepository {
	t.Helper()

	key := make([]byte, 32)
	_, err := rand.Read(key)
	require.NoError(t, err)
	cipher, err := security.NewAESTokenCipher(base64.StdEncoding.EncodeToString(key))
	require.NoError(t, err)
	return repositories.NewSQLiteDropboxLinkRepository(ds.DB, cipher)
}

func TestSQLiteDropboxLinkRepository_Lifecycle_PersistsEncryptedTokens(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	ds := newDatasource(t)
	users := repositories.NewSQLiteUserRepository(ds.DB)
	links := newDropboxRepository(t, ds)
	user := createUser(t, users, "a@b.com")
	expiresAt := time.Unix(time.Now().Add(time.Hour).Unix(), 0)
	account := models.ExternalAccount{Subject: "dbid:1", Email: "a@b.com"}

	// Act
	require.NoError(t, links.Save(ctx, models.NewDropboxLink(user.Id, account, "refresh-1", models.AccessToken{})))
	require.NoError(t, links.UpdateFolder(ctx, user.Id, "/Inkomi"))
	require.NoError(t, links.Save(ctx, models.NewDropboxLink(user.Id, account, "refresh-2", models.AccessToken{})))
	require.NoError(t, links.UpdateAccess(ctx, user.Id, models.AccessToken{Token: "access", ExpiresAt: expiresAt}))
	link, err := links.Get(ctx, user.Id)

	var stored string
	require.NoError(t, ds.DB.QueryRow(`SELECT refreshTokenEnc FROM dropbox_links`).Scan(&stored))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, &models.DropboxLink{
		UserId:       user.Id,
		AccountId:    "dbid:1",
		Email:        "a@b.com",
		RefreshToken: "refresh-2",
		Access:       models.AccessToken{Token: "access", ExpiresAt: expiresAt},
		Folder:       "/Inkomi",
	}, link)
	assert.NotContains(t, stored, "refresh")
}

func TestSQLiteDropboxLinkRepository_NotLinked_ReturnsErrDropboxNotLinked(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	ds := newDatasource(t)
	users := repositories.NewSQLiteUserRepository(ds.DB)
	links := newDropboxRepository(t, ds)
	user := createUser(t, users, "a@b.com")
	require.NoError(t, links.Save(ctx, models.NewDropboxLink(user.Id, models.ExternalAccount{Subject: "s"}, "r", models.AccessToken{})))

	// Act
	require.NoError(t, links.Delete(ctx, user.Id))
	_, getErr := links.Get(ctx, user.Id)
	folderErr := links.UpdateFolder(ctx, user.Id, "/x")
	accessErr := links.UpdateAccess(ctx, user.Id, models.AccessToken{Token: "a"})

	// Assert
	assert.ErrorIs(t, getErr, models.ErrDropboxNotLinked)
	assert.ErrorIs(t, folderErr, models.ErrDropboxNotLinked)
	assert.ErrorIs(t, accessErr, models.ErrDropboxNotLinked)
}

func TestSQLiteDatasource_ExecuteTx_Error_RollsBack(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	ds := newDatasource(t)
	users := repositories.NewSQLiteUserRepository(ds.DB)
	boom := errors.New("boom")

	// Act
	err := ds.ExecuteTx(ctx, func(ctx context.Context) error {
		require.NoError(t, users.Create(ctx, models.NewPasswordUser("a@b.com", "hash")))
		return boom
	})
	_, getErr := users.GetByEmail(ctx, "a@b.com")

	// Assert
	assert.ErrorIs(t, err, boom)
	assert.ErrorIs(t, getErr, models.ErrUserNotFound)
}
