package models

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	bracketsRe = regexp.MustCompile(`[\[(][^\])]*[\])]`)
	spacesRe   = regexp.MustCompile(`\s+`)
)

// CleanTitle turns the title or filename of a sent book into a searchable title: no extension,
// no bracketed tags ("(z-lib.org)", "[epub]") and no underscores.
func CleanTitle(title string) string {
	title = strings.TrimSpace(title)
	if ext := strings.ToLower(filepath.Ext(title)); ext == ".epub" || ext == ".pdf" || ext == ".kepub" {
		title = strings.TrimSuffix(title, filepath.Ext(title))
	}
	title = strings.TrimSuffix(title, ".kepub")
	title = bracketsRe.ReplaceAllString(title, " ")
	title = strings.NewReplacer("_", " ", "+", " ").Replace(title)
	return strings.TrimSpace(spacesRe.ReplaceAllString(title, " "))
}
