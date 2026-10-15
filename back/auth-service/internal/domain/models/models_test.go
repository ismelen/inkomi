package models_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/auth-service/internal/domain/models"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeEmail_Input_Result(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		input string
		email string
		err   error
	}{
		{"Valid", "user@example.com", "user@example.com", nil},
		{"Trimmed and lowercased", "  User@Example.COM ", "user@example.com", nil},
		{"Empty", "", "", models.ErrInvalidEmail},
		{"No at", "user.example.com", "", models.ErrInvalidEmail},
		{"With name", "User <user@example.com>", "", models.ErrInvalidEmail},
		{"Too long", strings.Repeat("a", 250) + "@example.com", "", models.ErrInvalidEmail},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			email, err := models.NormalizeEmail(tc.input)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.email, email)
		})
	}
}

func TestValidatePassword_Input_Result(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		password string
		err      error
	}{
		{"Valid", "12345678", nil},
		{"Too short", "1234567", models.ErrWeakPassword},
		{"Too long", strings.Repeat("a", 129), models.ErrWeakPassword},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			err := models.ValidatePassword(tc.password)

			// Assert
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestValidateDeviceId_Input_Result(t *testing.T) {
	t.Parallel()

	// Act & Assert
	assert.NoError(t, models.ValidateDeviceId("device-1"))
	assert.ErrorIs(t, models.ValidateDeviceId(""), models.ErrInvalidDeviceId)
	assert.ErrorIs(t, models.ValidateDeviceId(strings.Repeat("a", 129)), models.ErrInvalidDeviceId)
}

func TestNewSocialUser_Account_Email(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		account  models.ExternalAccount
		email    string
		verified bool
	}{
		{"Verified", models.ExternalAccount{Email: "User@Example.com", EmailVerified: true}, "user@example.com", true},
		{"Unverified", models.ExternalAccount{Email: "user@example.com"}, "", false},
		{"Invalid", models.ExternalAccount{Email: "nope", EmailVerified: true}, "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			user := models.NewSocialUser(tc.account)

			// Assert
			assert.Equal(t, tc.email, user.Email)
			assert.Equal(t, tc.verified, user.EmailVerified)
			assert.Equal(t, models.RoleUser, user.Role)
		})
	}
}

func TestUser_ClaimEmail_Verified_KeepsPassword(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		verified bool
		password string
	}{
		{"Verified keeps password", true, "hash"},
		{"Unverified drops password", false, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			user := &models.User{Email: "a@b.com", EmailVerified: tc.verified, PasswordHash: "hash"}

			// Act
			user.ClaimEmail()

			// Assert
			assert.True(t, user.EmailVerified)
			assert.Equal(t, tc.password, user.PasswordHash)
		})
	}
}

func TestNormalizeDropboxFolder_Input_Result(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		input  string
		folder string
		err    error
	}{
		{"Root", "", "", nil},
		{"Slash", "/", "", nil},
		{"Without slashes", "Inkomi", "/Inkomi", nil},
		{"Nested with trailing slash", "/Inkomi/Books/", "/Inkomi/Books", nil},
		{"Double slash", "/a//b", "", models.ErrInvalidDropboxFolder},
		{"Dot dot", "/a/../b", "", models.ErrInvalidDropboxFolder},
		{"Control char", "/a\nb", "", models.ErrInvalidDropboxFolder},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			folder, err := models.NormalizeDropboxFolder(tc.input)

			// Assert
			assert.ErrorIs(t, err, tc.err)
			assert.Equal(t, tc.folder, folder)
		})
	}
}

func TestDropboxLink_HasValidAccess_Expiry_Result(t *testing.T) {
	t.Parallel()

	now := time.Now()
	cases := []struct {
		name   string
		access models.AccessToken
		valid  bool
	}{
		{"No token", models.AccessToken{ExpiresAt: now.Add(time.Hour)}, false},
		{"Valid", models.AccessToken{Token: "t", ExpiresAt: now.Add(time.Hour)}, true},
		{"Inside margin", models.AccessToken{Token: "t", ExpiresAt: now.Add(time.Minute)}, false},
		{"Expired", models.AccessToken{Token: "t", ExpiresAt: now.Add(-time.Minute)}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			link := &models.DropboxLink{Access: tc.access}

			// Act & Assert
			assert.Equal(t, tc.valid, link.HasValidAccess(now, 5*time.Minute))
		})
	}
}
