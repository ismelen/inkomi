package ports

import (
	"ismelen/inkomi/internal/domain"
)

type LibgenMirror interface {
	Search(query string) ([]domain.Book, error)
	Download(req domain.LibgenDownloadRequestDTO) (*domain.LibgenDownloadResultDTO, error)
	GetURL() string
}
