package volumeBuilder

import (
	"ismelen/ermc/v2/manga"
	"mime/multipart"
)

type BuilderI interface {
	FromMultipart(files ...*multipart.FileHeader) ([]*manga.Volume, error)
	FromPaths(paths ...string) ([]*manga.Volume, error)
}
