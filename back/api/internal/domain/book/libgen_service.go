package book

// LibgenService is the port used by HTTP handlers to interact with libgen.
type LibgenService interface {
	Search(query, language string, formats []string) ([]Book, error)
	Download(req LibgenDownloadRequest) (*LibgenDownloadResult, error)
}
