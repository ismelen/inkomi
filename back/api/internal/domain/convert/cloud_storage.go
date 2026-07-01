package convert

// CloudStorage is the port implemented by infra/storage to upload files to a cloud provider.
type CloudStorage interface {
	Upload(path string) error
	Init(token, folder string)
}
