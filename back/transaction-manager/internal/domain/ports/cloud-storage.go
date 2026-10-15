package ports

type ICloudStorage interface {
	GetUrl(id string, ext string) (string, error)
	Check(id string, ext string) bool
}
