package openlibrary

import (
	"errors"
	"strings"
)

var ErrNotFound = errors.New("not found in open library")

// Work is a work of a search or subject listing; keys have no "/works/" or "/authors/" prefix.
// With a language in the query, Title and CoverId are those of its edition in that language.
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
	Editions         struct {
		Docs []struct {
			Title  string `json:"title"`
			CoverI int    `json:"cover_i"`
		} `json:"docs"`
	} `json:"editions"`
}

func (d searchDoc) work() (Work, bool) {
	if d.Key == "" || d.Title == "" {
		return Work{}, false
	}
	keys := make([]string, len(d.AuthorKey))
	for i, key := range d.AuthorKey {
		keys[i] = trimKey(key)
	}
	title, coverId := d.Title, d.CoverI
	if len(d.Editions.Docs) > 0 {
		edition := d.Editions.Docs[0]
		if edition.Title != "" {
			title = edition.Title
		}
		if edition.CoverI > 0 {
			coverId = edition.CoverI
		}
	}
	return Work{
		Key:              trimKey(d.Key),
		Title:            title,
		Authors:          d.AuthorName,
		AuthorKeys:       keys,
		Subjects:         d.Subject,
		Languages:        d.Language,
		FirstPublishYear: d.FirstPublishYear,
		CoverId:          coverId,
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
