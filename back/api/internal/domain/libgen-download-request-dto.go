package domain

type LibgenDownloadRequestDTO struct {
	DownloadURL string `json:"download_url"`
	Title       string `json:"title"`
	Extension   string `json:"extension"`
}
