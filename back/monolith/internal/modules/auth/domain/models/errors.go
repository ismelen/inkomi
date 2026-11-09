package models

import "errors"

var (
	ErrInvalidEmail         = errors.New("invalid email")
	ErrWeakPassword         = errors.New("password must have between 8 and 128 characters")
	ErrInvalidDeviceId      = errors.New("invalid device id")
	ErrInvalidDropboxFolder = errors.New("invalid dropbox folder")
	ErrEmailTaken           = errors.New("email already registered")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrInvalidRefreshToken  = errors.New("invalid or expired refresh token")
	ErrInvalidExternalToken = errors.New("invalid external token")
	ErrUserNotFound         = errors.New("user not found")
	ErrIdentityNotFound     = errors.New("identity not found")
	ErrDropboxNotLinked     = errors.New("dropbox account not linked")
	ErrDropboxAccountTaken  = errors.New("dropbox account linked to another user")
	ErrFCMTokenNotFound     = errors.New("fcm token not found")
	ErrInvalidFCMToken      = errors.New("invalid fcm token")
)
