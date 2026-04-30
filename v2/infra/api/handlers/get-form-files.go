package handlers

import (
	"fmt"
	"mime/multipart"

	"github.com/labstack/echo/v4"
)

func GetFormFiles(c echo.Context, key string) ([]*multipart.FileHeader, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err 
	}
	
	formFiles := form.File[key]
	if len(formFiles) == 0 {
		return nil,  fmt.Errorf("no files attached") 
	}

	return formFiles, nil
}