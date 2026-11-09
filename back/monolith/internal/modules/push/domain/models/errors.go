package models

import "errors"

var (
	ErrDropboxNotLinked = errors.New("dropbox account not linked")
	ErrFCMTokenNotFound = errors.New("fcm token not found")
	ErrUploadRejected   = errors.New("upload rejected by cloud drive")
)
