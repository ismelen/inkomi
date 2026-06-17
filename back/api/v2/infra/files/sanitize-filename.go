package files

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

func SanitizeFilename(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("empty filename")
	}

	sanitized := strings.Map(func(r rune) rune {
		if r == 0 || unicode.IsControl(r) {
			return -1
		}
		return r
	}, input)

	if sanitized == "" {
		return "", fmt.Errorf("invalid characters only")
	}

	cleaned := filepath.Clean(sanitized)
	cleaned = strings.TrimLeft(cleaned, "/\\")

	if strings.ContainsAny(cleaned, "/\\") || cleaned == "." || cleaned == ".." {
		return "", fmt.Errorf("invalid path")
	}

	cleaned = strings.Trim(cleaned, " .")
	cleaned = regexp.MustCompile(`\s{2,}`).ReplaceAllString(cleaned, " ")

	if len(cleaned) > 255 {
		ext := filepath.Ext(cleaned)
		base := strings.TrimSuffix(cleaned, ext)
		base = base[:255-len(ext)]
		base = strings.TrimRight(base, " .")
		cleaned = base + ext
	}

	if cleaned == "" {
		return "", fmt.Errorf("invalid filename")
	}

	return cleaned, nil
}