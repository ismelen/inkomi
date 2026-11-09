package dtos

type UploadRequestDTO struct {
	EReaderKey string               `json:"ereader_key"`
	Sources    []SourceDTO          `json:"sources"`
	Configs    map[string]ConfigDTO `json:"config"`
}
