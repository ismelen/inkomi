package epub

import (
	"bytes"
	"encoding/xml"
	"strings"
)

func xmlEscape(s string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

func sanitizeID(id string) string {
	if id == "" {
		return "item"
	}
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			b.WriteRune(r)
		default:
			b.WriteRune('_')
		}
	}
	return b.String()
}

func removeToken(s, token string) string {
	fields := strings.Fields(s)
	out := fields[:0]
	for _, f := range fields {
		if f != token {
			out = append(out, f)
		}
	}
	return strings.Join(out, " ")
}
