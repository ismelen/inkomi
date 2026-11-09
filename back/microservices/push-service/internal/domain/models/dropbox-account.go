package models

import "strings"

type DropboxAccount struct {
	Token  string
	Folder string
}

// NewDropboxAccount normalizes folder to "" (the root) or "/a/b" without a trailing slash.
func NewDropboxAccount(token, folder string) (*DropboxAccount, error) {
	if token == "" {
		return nil, ErrDropboxNotLinked
	}

	folder = strings.Trim(strings.TrimSpace(folder), "/")
	if folder != "" {
		folder = "/" + folder
	}
	return &DropboxAccount{Token: token, Folder: folder}, nil
}

func (d *DropboxAccount) Path(filename string) string {
	return d.Folder + "/" + filename
}
