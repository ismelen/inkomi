package ports

type ICloudStorage interface {
	// GetUrl returns a presigned PUT url for id/filename, the key every worker reads.
	GetUrl(id string, filename string) (string, error)
	Check(id string, filename string) bool
}
