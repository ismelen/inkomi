package handlers

import (
	"fmt"
	"mime/multipart"
	"net/http"
)

func GetFormFiles(r *http.Request, key string) ([]*multipart.FileHeader, error) {
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, fmt.Errorf("no files attached")
	}

	formFiles := r.MultipartForm.File[key]
	if len(formFiles) == 0 {
		return nil, fmt.Errorf("no files attached")
	}

	return formFiles, nil
}
