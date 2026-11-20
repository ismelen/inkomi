package models

import "time"

// SentBook is a book a user sent: a root file (a library book or an epub) that reached sent or done.
type SentBook struct {
	SourceId  string
	UserId    int
	Title     string
	Filename  string
	UpdatedAt time.Time
}
