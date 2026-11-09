package dtos

// RetryUploadResponse is the new presigned PUT url of a file to upload again.
type RetryUploadResponse struct {
	Id  string `json:"id"`
	Url string `json:"url"`
}
