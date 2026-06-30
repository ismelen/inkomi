package book

import "io"

type LibgenDownloadRequest struct {
	DownloadURL string
	Title       string
	Extension   string
}

type LibgenDownloadResult struct {
	Stream        io.ReadCloser
	ContentType   string
	ContentLength int64
	Filename      string
}
