package helpers

import (
	"regexp"
	"strconv"
)

var numRe = regexp.MustCompile(`\d+|\D+`)

func AlphanumericCmp(a, b string) bool {
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
