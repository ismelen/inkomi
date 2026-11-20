package strutil

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Normalize lowercases value, strips accents, turns spaces and underscores into dashes and
// drops anything that isn't a letter, a number, a dash or a dot.
func Normalize(value string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, value)
	if err != nil {
		result = value
	}

	result = strings.ToLower(result)
	result = strings.NewReplacer(" ", "-", "_", "-").Replace(result)
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '-' || r == '.' {
			return r
		}
		return -1
	}, result)
}
