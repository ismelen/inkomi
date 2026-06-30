package book

// LibgenMirror is the port implemented by each concrete mirror (classic, plus, slum...).
type LibgenMirror interface {
	Search(query string) ([]Book, error)
	Download(req LibgenDownloadRequest) (*LibgenDownloadResult, error)
	GetURL() string
}
