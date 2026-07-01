package book

import "io"

type LibgenDownload struct {
	Stream        io.ReadCloser
	ContentType   string
	ContentLength int64
	Filename      string
}
