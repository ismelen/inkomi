package volumeBuilder

import (
	"ismelen/ermc/internal/domain"
	"mime/multipart"
	"os"
)

type BuilderI interface {
	FromMultipart(settings *domain.Settings, files ...*multipart.FileHeader) ([]*domain.Volume, error)
	FromPaths(settings *domain.Settings, files ...os.DirEntry) ([]*domain.Volume, error)
}
