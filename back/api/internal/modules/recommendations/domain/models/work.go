package models

import "strings"

const (
	passageSubjects    = 10
	passageDescription = 600
)

type Origin string

const (
	// OriginUser is a work some user sent or wished.
	OriginUser Origin = "user"
	// OriginCandidate is a work found through the users' authors and subjects.
	OriginCandidate Origin = "candidate"
)

// Work is a book of the catalog, identified by its Open Library work key (OL…W).
type Work struct {
	Key              string
	Title            string
	Authors          []string
	AuthorKeys       []string
	Subjects         []string
	Languages        []string
	FirstPublishYear int
	CoverId          int
	Description      string
	Origin           Origin
}

// Passage is the text the work is embedded from: what it is about, without noise.
func (w *Work) Passage() string {
	parts := []string{strings.TrimSpace(w.Title)}
	if len(w.Authors) > 0 {
		parts = append(parts, strings.Join(w.Authors, ", "))
	}
	if subjects := w.Subjects[:min(len(w.Subjects), passageSubjects)]; len(subjects) > 0 {
		parts = append(parts, strings.Join(subjects, ", "))
	}
	if description := strings.TrimSpace(w.Description); description != "" {
		if runes := []rune(description); len(runes) > passageDescription {
			description = string(runes[:passageDescription])
		}
		parts = append(parts, description)
	}
	return strings.Join(parts, " · ")
}

// HasAuthorIn reports whether any author of the work is in authors.
func (w *Work) HasAuthorIn(authors map[string]bool) bool {
	for _, key := range w.AuthorKeys {
		if authors[key] {
			return true
		}
	}
	return false
}
