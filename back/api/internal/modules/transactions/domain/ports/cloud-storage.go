package ports

import "time"

type ICloudStorage interface {
	// GetUrl returns a presigned PUT url for id/filename, the key every worker reads. size is the
	// declared size of the file; storages that reserve space reject bigger bodies.
	GetUrl(id string, filename string, size int64) (string, error)
	// GetDownloadUrl returns a presigned GET url for id/filename that downloads it as an attachment, and when it expires.
	GetDownloadUrl(id string, filename string) (string, time.Time, error)
	Check(id string, filename string) bool
}
