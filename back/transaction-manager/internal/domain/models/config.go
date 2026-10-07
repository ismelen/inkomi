package models

import "time"

type Config struct {
	Hash      string    `json:"hash"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"createdAt,omitempty"`
	LastUsed  time.Time `json:"lastUsed,omitempty"`
}
