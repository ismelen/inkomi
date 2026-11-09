package models

import (
	"fmt"
	"regexp"
	"strings"
)

var md5Re = regexp.MustCompile(`^[0-9a-f]{32}$`)

// MD5 is the lowercase hex md5 that identifies a book file in libgen.
type MD5 string

func NewMD5(value string) (MD5, error) {
	md5 := strings.ToLower(strings.TrimSpace(value))
	if !md5Re.MatchString(md5) {
		return "", fmt.Errorf("%q: %w", value, ErrInvalidMD5)
	}
	return MD5(md5), nil
}
