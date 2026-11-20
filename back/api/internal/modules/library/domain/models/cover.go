package models

import "io"

// Cover is an open download of a book cover image; Body must be closed by the caller.
type Cover struct {
	Body        io.ReadCloser
	ContentType string
}
