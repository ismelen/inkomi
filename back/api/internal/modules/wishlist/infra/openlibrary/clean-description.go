package openlibrary

import (
	"regexp"
	"strings"
)

var (
	// a line of dashes starts the "See also" / "Contained in" link lists
	separatorRe    = regexp.MustCompile(`(?m)^\s*-{3,}\s*$`)
	referenceRe    = regexp.MustCompile(`(?m)^\s*\[[^\]]+\]:\s*\S+\s*$`)
	sourceRe       = regexp.MustCompile(`\(\s*\[(?i:source)\]\[[^\]]*\]\s*\)`)
	inlineLinkRe   = regexp.MustCompile(`\[([^\]]+)\]\([^)]*\)`)
	refLinkRe      = regexp.MustCompile(`\[([^\]]+)\]\[[^\]]*\]`)
	blankLinesRe   = regexp.MustCompile(`\n{3,}`)
	trailingSpaces = regexp.MustCompile(`[ \t]+\n`)
)

// cleanDescription turns the markdown of an Open Library description into plain text: links
// keep their text, and reference lists, source notes and the trailing link sections go.
func cleanDescription(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	if loc := separatorRe.FindStringIndex(text); loc != nil {
		text = text[:loc[0]]
	}
	text = referenceRe.ReplaceAllString(text, "")
	text = sourceRe.ReplaceAllString(text, "")
	text = inlineLinkRe.ReplaceAllString(text, "$1")
	text = refLinkRe.ReplaceAllString(text, "$1")
	text = trailingSpaces.ReplaceAllString(text, "\n")
	text = blankLinesRe.ReplaceAllString(text, "\n\n")
	return strings.TrimSpace(text)
}
