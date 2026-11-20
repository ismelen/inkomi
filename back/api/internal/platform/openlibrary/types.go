package openlibrary

import (
	"errors"
	"strings"
)

var ErrNotFound = errors.New("not found in open library")

// Work is a work of a search or subject listing; keys have no "/works/" or "/authors/" prefix.
type Work struct {
	Key              string
	Title            string
	Authors          []string
	AuthorKeys       []string
	Subjects         []string
	Languages        []string
	FirstPublishYear int
	CoverId          int
}

type Author struct {
	Key       string
	Name      string
	TopWork   string
	WorkCount int
}

type searchDoc struct {
	Key              string   `json:"key"`
	Title            string   `json:"title"`
	AuthorName       []string `json:"author_name"`
	AuthorKey        []string `json:"author_key"`
	Subject          []string `json:"subject"`
	FirstPublishYear int      `json:"first_publish_year"`
	Language         []string `json:"language"`
	CoverI           int      `json:"cover_i"`
}

func (d searchDoc) work() (Work, bool) {
	if d.Key == "" || d.Title == "" {
		return Work{}, false
	}
	keys := make([]string, len(d.AuthorKey))
	for i, key := range d.AuthorKey {
		keys[i] = trimKey(key)
	}
	return Work{
		Key:              trimKey(d.Key),
		Title:            d.Title,
		Authors:          d.AuthorName,
		AuthorKeys:       keys,
		Subjects:         d.Subject,
		Languages:        d.Language,
		FirstPublishYear: d.FirstPublishYear,
		CoverId:          d.CoverI,
	}, true
}

// trimKey turns "/works/OL1W" or "/authors/OL1A" into "OL1W" or "OL1A".
func trimKey(key string) string {
	key = strings.TrimSpace(key)
	if i := strings.LastIndex(key, "/"); i >= 0 {
		return key[i+1:]
	}
	return key
}
