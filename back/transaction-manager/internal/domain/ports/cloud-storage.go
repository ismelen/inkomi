package ports

type CloudStorage interface {
	GetUrl(id string, ext string) (string, error)
}
