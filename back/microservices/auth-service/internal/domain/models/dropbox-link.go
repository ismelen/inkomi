package models

import (
	"strings"
	"time"
	"unicode"
)

const maxDropboxFolderLength = 1024

// DropboxLink is the Dropbox account where the files of a user are uploaded.
type DropboxLink struct {
	UserId       int
	AccountId    string
	Email        string
	RefreshToken string
	Access       AccessToken
	Folder       string
}

func NewDropboxLink(userId int, account ExternalAccount, refreshToken string, access AccessToken) *DropboxLink {
	return &DropboxLink{
		UserId:       userId,
		AccountId:    account.Subject,
		Email:        account.Email,
		RefreshToken: refreshToken,
		Access:       access,
	}
}

// HasValidAccess reports whether the cached access token is still valid after margin.
func (d *DropboxLink) HasValidAccess(now time.Time, margin time.Duration) bool {
	return d.Access.Token != "" && d.Access.ExpiresAt.After(now.Add(margin))
}

// NormalizeDropboxFolder returns "" for the root or "/a/b" without a trailing slash.
func NormalizeDropboxFolder(folder string) (string, error) {
	folder = strings.Trim(strings.TrimSpace(folder), "/")
	if len(folder) > maxDropboxFolderLength || strings.ContainsFunc(folder, unicode.IsControl) {
		return "", ErrInvalidDropboxFolder
	}
	for _, part := range strings.Split(folder, "/") {
		if folder != "" && (part == "" || part == "." || part == "..") {
			return "", ErrInvalidDropboxFolder
		}
	}

	if folder == "" {
		return "", nil
	}
	return "/" + folder, nil
}
