package models

import "time"

type SourceDownload struct {
	Url       string
	Filename  string
	Size      *int64
	ExpiresAt time.Time
}
