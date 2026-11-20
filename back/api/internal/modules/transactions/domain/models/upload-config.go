package models

// UploadConfig is a manga config of an upload request. Data is the serialized config the manga
// module reads; Hash is filled when it is saved.
type UploadConfig struct {
	Data     string
	Kepubify bool
	Hash     string
}
