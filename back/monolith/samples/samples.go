// Package samples embeds the preview samples, which the monolith seeds into its local storage
// (with R2 the e2e image uploads them).
package samples

import "embed"

//go:embed *.jpg
var Files embed.FS

const (
	Color = "color.jpg"
	BW    = "bw.jpg"
)
