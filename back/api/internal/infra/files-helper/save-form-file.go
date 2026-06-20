package filesHelper

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

func CopyFile(src string, outBase string) (string, error) {
	file, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer file.Close()

	filename := filepath.Base(src)
	dstPath := filepath.Join(outBase, filename)
	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	return dstPath, err
}

func CopyFormFiles(files []*multipart.FileHeader, outBase string) (string, []string, error) {
	basePath, err := CreateSanitizedFolder("", outBase)
	if err != nil {
		return "", nil, err
	}

	paths := make([]string, 0, len(files))
	for _, file := range files {
		path := filepath.Join(basePath, file.Filename)
		if err := copyFormFile(file, path); err != nil {
			return "", nil, err
		}
		paths = append(paths, path)
	}

	return basePath, paths, nil
}

func copyFormFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
