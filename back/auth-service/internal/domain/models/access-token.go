package models

import "time"

type AccessToken struct {
	Token     string
	ExpiresAt time.Time
}
