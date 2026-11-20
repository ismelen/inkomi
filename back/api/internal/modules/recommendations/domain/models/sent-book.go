package models

import "time"

// SentBook is a book a user sent, as the transactions module reports it.
type SentBook struct {
	SourceId  string
	UserId    int
	Title     string
	Filename  string
	UpdatedAt time.Time
}
