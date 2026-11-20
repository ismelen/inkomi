package models

// DropboxFolder is a folder of the user's Dropbox; Path is "/a/b" as Dropbox displays it.
type DropboxFolder struct {
	Name string
	Path string
}
