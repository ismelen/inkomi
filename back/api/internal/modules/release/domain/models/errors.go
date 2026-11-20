package models

import "errors"

var (
	ErrReleaseNotFound   = errors.New("no release available")
	ErrAPKNotFound       = errors.New("release has no apk")
	ErrInvalidVersion    = errors.New("invalid release version")
	ErrSourceUnavailable = errors.New("release source unavailable")
)
