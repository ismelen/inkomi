package models

import "time"

// Session is the login of a user on a device; only the hash of its refresh token is stored.
type Session struct {
	UserId      int
	DeviceId    string
	RefreshHash string
	ExpiresAt   time.Time
}
