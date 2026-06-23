package domain

import "io"

type LibgenDownloadResultDTO struct {
	Stream        io.ReadCloser
	ContentType   string
	ContentLength int64
	Filename      string
}
