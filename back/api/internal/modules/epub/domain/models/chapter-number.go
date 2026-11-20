package models

import (
	"regexp"
	"strconv"
)

var chapterPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bch(?:apter)?\.?\s*(\d+(?:\.\d+)?)`),
	regexp.MustCompile(`(?i)\bcap[ií]?tulo\.?\s*(\d+(?:\.\d+)?)|(?i)\bcap\.?\s*(\d+(?:\.\d+)?)`),
	regexp.MustCompile(`(?i)\bepisodi?o?\.?\s*(\d+(?:\.\d+)?)`),
	regexp.MustCompile(`(?:^|[\/])#(\d+(?:\.\d+)?)\s*-`),
}

var leadingChapterPattern = regexp.MustCompile(`^(\d+(?:\.\d+)?)---`)

var fallbackNumberPattern = regexp.MustCompile(`\d+(?:\.\d+)?`)

var noisePattern = regexp.MustCompile(`(?i)\b(19|20)\d{2}\b|\b\d{3,4}p\b|\bvol(?:ume)?\.?\s*\d+`)

var hashSuffixPattern = regexp.MustCompile(`(?i)[0-9a-f]{4,8}$`)

// ExtractChapterNumber finds the chapter number in a title or filename.
// Explicit markers (Ch., Capítulo, Episode, #N -) win over the last loose number.
func ExtractChapterNumber(name string) (float64, bool) {
	if m := leadingChapterPattern.FindStringSubmatch(name); m != nil {
		if n, err := strconv.ParseFloat(m[1], 64); err == nil {
			return n, true
		}
	}

	for _, re := range chapterPatterns {
		m := re.FindStringSubmatch(name)
		if m == nil {
			continue
		}
		for _, g := range m[1:] {
			if g == "" {
				continue
			}
			if n, err := strconv.ParseFloat(g, 64); err == nil {
				return n, true
			}
		}
	}

	cleaned := noisePattern.ReplaceAllString(name, "")
	cleaned = hashSuffixPattern.ReplaceAllString(cleaned, "")
	matches := fallbackNumberPattern.FindAllString(cleaned, -1)
	if len(matches) == 0 {
		return 0, false
	}

	n, err := strconv.ParseFloat(matches[len(matches)-1], 64)
	return n, err == nil
}
