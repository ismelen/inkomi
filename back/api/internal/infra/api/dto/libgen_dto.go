package dto

// LibgenDownloadRequest is the HTTP JSON body for a book download request.
type LibgenDownloadRequest struct {
	DownloadURL string `json:"download_url"`
	Title       string `json:"title"`
	Extension   string `json:"extension"`
}
