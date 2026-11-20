package models

// UploadRequest is a batch of sources a user wants to send, with the manga configs they refer to
// by key.
type UploadRequest struct {
	EReaderKey string
	Sources    []UploadSource
	Configs    map[string]UploadConfig
}
