package getdropboxaccess_test

import (
	"context"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/getdropboxaccess"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	links     *mocks.MockDropboxLinkRepository
	oauth     *mocks.MockDropboxOAuth
	refreshes int
	updated   []models.AccessToken
	deleted   []int
}

func newFixture(access models.AccessToken) *fixture {
	f := &fixture{}
	f.links = &mocks.MockDropboxLinkRepository{
		GetFn: func(ctx context.Context, userId int) (*models.DropboxLink, error) {
			return &models.DropboxLink{UserId: userId, RefreshToken: "dbx-refresh", Access: access, Folder: "/Inkomi"}, nil
		},
		UpdateAccessFn: func(ctx context.Context, userId int, access models.AccessToken) error {
			f.updated = append(f.updated, access)
			return nil
		},
		DeleteFn: func(ctx context.Context, userId int) error {
			f.deleted = append(f.deleted, userId)
			return nil
		},
	}
	f.oauth = &mocks.MockDropboxOAuth{
		RefreshFn: func(ctx context.Context, refreshToken string) (models.AccessToken, error) {
			f.refreshes++
			return models.AccessToken{Token: "fresh", ExpiresAt: time.Now().Add(4 * time.Hour)}, nil
		},
	}
	return f
}

func (f *fixture) execute() (*models.DropboxLink, error) {
	return getdropboxaccess.NewGetDropboxAccessUC(f.links, f.oauth).Execute(context.Background(), 2)
}

func TestGetDropboxAccessUC_Execute_CachedAccess_Result(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		access    models.AccessToken
		token     string
		refreshes int
	}{
		{"Valid cache", models.AccessToken{Token: "cached", ExpiresAt: time.Now().Add(time.Hour)}, "cached", 0},
		{"About to expire", models.AccessToken{Token: "cached", ExpiresAt: time.Now().Add(time.Minute)}, "fresh", 1},
		{"No cache", models.AccessToken{}, "fresh", 1},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture(tc.access)

			// Act
			link, err := f.execute()

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.token, link.Access.Token)
			assert.Equal(t, "/Inkomi", link.Folder)
			assert.Equal(t, tc.refreshes, f.refreshes)
			assert.Len(t, f.updated, tc.refreshes)
		})
	}
}

func TestGetDropboxAccessUC_Execute_RevokedToken_UnlinksAndReturnsErrDropboxNotLinked(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(models.AccessToken{})
	f.oauth.RefreshFn = nil

	// Act
	link, err := f.execute()

	// Assert
	assert.ErrorIs(t, err, models.ErrDropboxNotLinked)
	assert.Nil(t, link)
	assert.Equal(t, []int{2}, f.deleted)
}

func TestGetDropboxAccessUC_Execute_NotLinked_ReturnsErrDropboxNotLinked(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(models.AccessToken{})
	f.links.GetFn = nil

	// Act
	link, err := f.execute()

	// Assert
	assert.ErrorIs(t, err, models.ErrDropboxNotLinked)
	assert.Nil(t, link)
}
