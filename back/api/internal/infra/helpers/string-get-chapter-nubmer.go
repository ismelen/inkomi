package helpers

import (
	"regexp"
	"strconv"
)

var chapterPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bch(?:apter)?\.?\s*(\d+(?:\.\d+)?)`),
	regexp.MustCompile(`(?i)\bcap[ií]?tulo\.?\s*(\d+(?:\.\d+)?)|(?i)\bcap\.?\s*(\d+(?:\.\d+)?)`),
	regexp.MustCompile(`(?i)\bepisodi?o?\.?\s*(\d+(?:\.\d+)?)`),
}

func ExtractChapterNumber(filename string) (float64, bool) {
	for _, re := range chapterPatterns {
		m := re.FindStringSubmatch(filename)
		if m == nil {
			continue
		}
		for _, g := range m[1:] {
			if g != "" {
				n, err := strconv.ParseFloat(g, 64)
				if err == nil {
					return n, true
				}
			}
		}
	}
	return 0, false
}
