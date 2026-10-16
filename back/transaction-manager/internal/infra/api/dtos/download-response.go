package dtos

import (
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
)

type DownloadResponse struct {
	Url       string    `json:"url"`
	Filename  string    `json:"filename"`
	Size      *int64    `json:"size"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func NewDownloadResponse(d *models.SourceDownload) *DownloadResponse {
	return &DownloadResponse{
		Url:       d.Url,
		Filename:  d.Filename,
		Size:      d.Size,
		ExpiresAt: d.ExpiresAt.UTC(),
	}
}
