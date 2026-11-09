package models

import (
	"fmt"
	"strconv"
	"strings"
)

// Version is MAJOR.MINOR.PATCH, the expo.version of front/mobile/app.json.
type Version struct {
	Major, Minor, Patch int
}

// ParseVersion accepts the release tag, with or without the leading "v".
func ParseVersion(raw string) (Version, error) {
	parts := strings.Split(strings.TrimPrefix(strings.TrimSpace(raw), "v"), ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("%w: %q", ErrInvalidVersion, raw)
	}

	numbers := make([]int, 3)
	for i, part := range parts {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return Version{}, fmt.Errorf("%w: %q", ErrInvalidVersion, raw)
		}
		numbers[i] = n
	}

	return Version{Major: numbers[0], Minor: numbers[1], Patch: numbers[2]}, nil
}

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}
