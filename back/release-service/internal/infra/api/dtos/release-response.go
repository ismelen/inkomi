package dtos

import (
	"time"

	"github.com/ismelen/inkomi/back/release-service/internal/domain/models"
)

type ReleaseResponse struct {
	Version     string    `json:"version"`
	Notes       string    `json:"notes"`
	PublishedAt time.Time `json:"publishedAt"`
	FileName    string    `json:"fileName"`
	Size        int64     `json:"size"`
	SHA256      string    `json:"sha256,omitempty"`
	// DownloadPath is relative to the API base url.
	DownloadPath string `json:"downloadPath"`
}

func NewReleaseResponse(release *models.Release, downloadPath string) *ReleaseResponse {
	return &ReleaseResponse{
		Version:      release.Version.String(),
		Notes:        release.Notes,
		PublishedAt:  release.PublishedAt.UTC(),
		FileName:     release.APK.Name,
		Size:         release.APK.Size,
		SHA256:       release.APK.SHA256,
		DownloadPath: downloadPath,
	}
}
