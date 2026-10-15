package socialsignin_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/mocks"
	"github.com/ismelen/inkomi/back/auth-service/internal/usecases/socialsignin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fixture struct {
	users      *mocks.MockUserRepository
	identities *mocks.MockIdentityRepository
	byEmail    *models.User
	created    *models.User
	updated    *models.User
	linked     []models.Identity
}

func newFixture(byEmail *models.User) *fixture {
	f := &fixture{byEmail: byEmail}
	f.users = &mocks.MockUserRepository{
		GetByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			if f.byEmail == nil || f.byEmail.Email != email {
				return nil, models.ErrUserNotFound
			}
			return f.byEmail, nil
		},
		GetByIdFn: func(ctx context.Context, id int) (*models.User, error) {
			return &models.User{Id: id}, nil
		},
		CreateFn: func(ctx context.Context, user *models.User) error {
			user.Id = 99
			f.created = user
			return nil
		},
		UpdateFn: func(ctx context.Context, user *models.User) error {
			f.updated = user
			return nil
		},
	}
	f.identities = &mocks.MockIdentityRepository{
		CreateFn: func(ctx context.Context, identity models.Identity) error {
			f.linked = append(f.linked, identity)
			return nil
		},
	}
	return f
}

func (f *fixture) execute(account models.ExternalAccount) (*models.User, error) {
	uc := socialsignin.NewSocialSignInUC(&mocks.MockTxManager{}, f.users, f.identities)
	return uc.Execute(context.Background(), account)
}

func googleAccount(email string, verified bool) models.ExternalAccount {
	return models.ExternalAccount{Provider: models.ProviderGoogle, Subject: "g-1", Email: email, EmailVerified: verified}
}

func TestSocialSignInUC_Execute_KnownIdentity_ReturnsItsUser(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(nil)
	f.identities.GetFn = func(ctx context.Context, provider models.Provider, subject string) (*models.Identity, error) {
		return &models.Identity{Provider: provider, Subject: subject, UserId: 5}, nil
	}

	// Act
	user, err := f.execute(googleAccount("a@b.com", true))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 5, user.Id)
	assert.Nil(t, f.created)
	assert.Empty(t, f.linked)
}

func TestSocialSignInUC_Execute_VerifiedEmailOfExistingUser_LinksIdentity(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		existing *models.User
		password string
	}{
		{"Verified account keeps password", &models.User{Id: 2, Email: "a@b.com", EmailVerified: true, PasswordHash: "h"}, "h"},
		{"Unverified account loses password", &models.User{Id: 2, Email: "a@b.com", PasswordHash: "h"}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture(tc.existing)

			// Act
			user, err := f.execute(googleAccount("A@b.com", true))

			// Assert
			require.NoError(t, err)
			assert.Equal(t, 2, user.Id)
			require.NotNil(t, f.updated)
			assert.True(t, f.updated.EmailVerified)
			assert.Equal(t, tc.password, f.updated.PasswordHash)
			assert.Nil(t, f.created)
			assert.Equal(t, []models.Identity{{Provider: models.ProviderGoogle, Subject: "g-1", UserId: 2}}, f.linked)
		})
	}
}

func TestSocialSignInUC_Execute_NoMatchingUser_CreatesUser(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		account models.ExternalAccount
		email   string
	}{
		{"Unverified email of existing user", googleAccount("a@b.com", false), ""},
		{"New verified email", googleAccount("new@b.com", true), "new@b.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			f := newFixture(&models.User{Id: 2, Email: "a@b.com"})

			// Act
			user, err := f.execute(tc.account)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, 99, user.Id)
			require.NotNil(t, f.created)
			assert.Equal(t, tc.email, f.created.Email)
			assert.Nil(t, f.updated)
			assert.Equal(t, []models.Identity{{Provider: models.ProviderGoogle, Subject: "g-1", UserId: 99}}, f.linked)
		})
	}
}

func TestSocialSignInUC_Execute_RepositoryFails_ReturnsError(t *testing.T) {
	t.Parallel()

	// Arrange
	boom := errors.New("boom")
	f := newFixture(nil)
	f.identities.CreateFn = func(ctx context.Context, identity models.Identity) error { return boom }

	// Act
	user, err := f.execute(googleAccount("a@b.com", true))

	// Assert
	assert.ErrorIs(t, err, boom)
	assert.Nil(t, user)
}
