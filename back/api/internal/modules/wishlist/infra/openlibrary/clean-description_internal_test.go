package openlibrary

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCleanDescription_Markdown_ReturnsPlainText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   string
		want string
	}{
		{"Plain", "  Ash falls from the sky.  ", "Ash falls from the sky."},
		{"Inline link", "Set in [Scadrial](https://example.com/s) long ago.", "Set in Scadrial long ago."},
		{"Source and references", "A heist.\r\n\r\n([source][1])\r\n\r\n[1]: https://example.com", "A heist."},
		{"Trailing links", "A heist.\n\n----------\nAlso contained in:\n\n - [Omnibus](/works/OL1W)", "A heist."},
		{"Blank lines", "One.\n\n\n\nTwo.", "One.\n\nTwo."},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			got := cleanDescription(tc.in)

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}
