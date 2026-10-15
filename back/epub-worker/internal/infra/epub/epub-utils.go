package epub

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"slices"
	"strings"
)

func xmlEscape(s string) string {
	var buf bytes.Buffer
	// writing to a bytes.Buffer never fails
	_ = xml.EscapeText(&buf, []byte(s))
	return buf.String()
}

func sanitizeId(id string) string {
	if id == "" {
		return "item"
	}
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_', r == '.':
			return r
		default:
			return '_'
		}
	}, id)
}

func hasToken(s, token string) bool {
	return slices.Contains(strings.Fields(s), token)
}

func removeToken(s, token string) string {
	return strings.Join(slices.DeleteFunc(strings.Fields(s), func(f string) bool { return f == token }), " ")
}

func addToken(s, token string) string {
	return strings.TrimSpace(s + " " + token)
}

func errorAndClose(err error, c io.Closer) error {
	return errors.Join(err, c.Close())
}
