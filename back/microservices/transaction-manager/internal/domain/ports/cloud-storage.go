package ports

import "time"

type ICloudStorage interface {
	// GetUrl returns a presigned PUT url for id/filename, the key every worker reads.
	GetUrl(id string, filename string) (string, error)
	// GetDownloadUrl returns a presigned GET url for id/filename that downloads it as an attachment, and when it expires.
	GetDownloadUrl(id string, filename string) (string, time.Time, error)
	Check(id string, filename string) bool
}
