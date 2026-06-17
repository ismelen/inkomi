package volumeBuilder

import (
	"ismelen/ermc/internal/domain"
	"ismelen/ermc/internal/pkg"
	"mime/multipart"
)

type BuilderI interface {
	FromMultipart(settings *domain.Settings, files ...*multipart.FileHeader) ([]*domain.Volume, error)
	FromPaths(settings *domain.Settings, files ...pkg.Pair[string, int64]) ([]*domain.Volume, error)
}
