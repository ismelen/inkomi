package models

// UploadSource is a root of an upload request: a file, a library book or a folder with its items.
type UploadSource struct {
	UploadItem
	ShouldJoin bool
	Items      []UploadItem
}
