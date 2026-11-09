package models

type Profile struct {
	User      *User
	Providers []Provider
	// Dropbox is nil when the user has no Dropbox account linked.
	Dropbox *DropboxLink
}
