package strutil

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

var numRe = regexp.MustCompile(`\d+|\D+`)

func AlphanumericCmp(a, b string) bool {
	aNum, aOk := ExtractChapterNumber(a)
	bNum, bOk := ExtractChapterNumber(b)

	if aOk && bOk && aNum != bNum {
		return aNum < bNum
	}

	return alphanumericCmp(a, b)
}

func alphanumericCmp(a, b string) bool {
	as := numRe.FindAllString(a, -1)
	bs := numRe.FindAllString(b, -1)

	for i := 0; i < len(as) && i < len(bs); i++ {
		if as[i] == bs[i] {
			continue
		}

		an, aErr := strconv.Atoi(as[i])
		bn, bErr := strconv.Atoi(bs[i])

		if aErr == nil && bErr == nil {
			return an < bn
		}
		return as[i] < bs[i]
	}

	return len(as) < len(bs)
}
