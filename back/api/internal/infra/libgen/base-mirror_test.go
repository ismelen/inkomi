package libgen

import "testing"

func TestStripLibgenMetadata(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "real world case with Series and Author",
			input: "El Héroe de las Eras Series: Nacidos de la bruma 3 Author(s): Brandon Sanderson Publisher: Epublibre Year: 2008;2007 ISBN:",
			want:  "El Héroe de las Eras",
		},
		{
			name:  "clean title without metadata",
			input: "Dune",
			want:  "Dune",
		},
		{
			name:  "title with Authors suffix",
			input: "Clean Code Authors: Robert C. Martin",
			want:  "Clean Code",
		},
		{
			name:  "title with Publisher suffix",
			input: "The Hobbit Publisher: HarperCollins Year: 1937",
			want:  "The Hobbit",
		},
		{
			name:  "title with trailing whitespace",
			input: "  My Book  Author(s): Someone ",
			want:  "My Book",
		},
		{
			name:  "title with Edition suffix",
			input: "Go Programming Language Edition: 1st",
			want:  "Go Programming Language",
		},
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripLibgenMetadata(tt.input)
			if got != tt.want {
				t.Errorf("stripLibgenMetadata(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
