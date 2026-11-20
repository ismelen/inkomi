package models

// UploadItem is one source of an upload request. Id and Url are filled when it is saved.
type UploadItem struct {
	Title            string
	Type             SourceType
	Size             *int64
	Filename         string
	ConfigKey        *string
	ReadingDirection ReadingDirection
	ToCloud          bool
	Kepubify         bool
	Url              string
	Id               string
}
