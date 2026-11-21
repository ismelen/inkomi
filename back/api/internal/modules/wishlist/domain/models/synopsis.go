package models

import (
	"regexp"
	"time"
)

var languageRe = regexp.MustCompile(`^[a-z]{2}$`)

// Synopsis is the synopsis of a work in a language (ISO 639-1). An empty Text records that the
// catalog had none when it was fetched.
type Synopsis struct {
	WorkKey   string
	Language  string
	Text      string
	FetchedAt time.Time
}

// IsLanguage reports whether code looks like an ISO 639-1 code.
func IsLanguage(code string) bool {
	return languageRe.MatchString(code)
}
