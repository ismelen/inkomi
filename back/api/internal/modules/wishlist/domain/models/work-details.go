package models

import "regexp"

var workKeyRe = regexp.MustCompile(`^OL\d+W$`)

// WorkDetails is what the catalog knows of a work. Description may be empty.
type WorkDetails struct {
	WorkKey        string
	Title          string
	Description    string
	CoverURL       string
	Subjects       []string
	FirstPublished string
}

// IsWorkKey reports whether key looks like an Open Library work key (OL…W).
func IsWorkKey(key string) bool {
	return workKeyRe.MatchString(key)
}

const catalogCoverPrefix = "https://covers.openlibrary.org/"

// IsCatalogCoverURL reports whether url is a cover of the catalog, the only ones stored.
func IsCatalogCoverURL(url string) bool {
	return len(url) > len(catalogCoverPrefix) && url[:len(catalogCoverPrefix)] == catalogCoverPrefix
}
