package models

import (
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/shared/strutil"
)

// TitleWords are the normalized words of a title or a name (no accents, no case), skipping
// one-letter ones.
func TitleWords(value string) []string {
	var words []string
	for _, word := range strings.FieldsFunc(strutil.Normalize(value), func(r rune) bool { return r == '-' || r == '.' }) {
		if len([]rune(word)) > 1 {
			words = append(words, word)
		}
	}
	return words
}

// Overlap is the share of the words of candidate that appear in query.
func Overlap(candidate, query []string) float64 {
	if len(candidate) == 0 {
		return 0
	}
	present := make(map[string]bool, len(query))
	for _, word := range query {
		present[word] = true
	}
	found := 0
	for _, word := range candidate {
		if present[word] {
			found++
		}
	}
	return float64(found) / float64(len(candidate))
}
