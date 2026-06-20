package filesHelper

import (
	"archive/zip"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func UnzipFormZip(src *multipart.FileHeader, outBase string) (string, []string, error) {
	file, err := src.Open()
	if err != nil {
		return "", nil, err
	}
	defer file.Close()

	reader, err := zip.NewReader(file, src.Size)
	if err != nil {
		return "", nil, err
	}

	basePath, err := CreateSanitizedFolder(src.Filename, outBase)
	if err != nil {
		return "", nil, err
	}

	paths := make([]string, 0, len(reader.File)-1)
	for _, file := range reader.File[1:] {
		path := filepath.Join(basePath, file.FileInfo().Name())
		if err := ExtractFile(file, path); err != nil {
			return "", nil, err
		}
		paths = append(paths, path)
	}

	return basePath, paths, nil
}

func UnzipFile(src string, outBase string) (string, string, []string, error) {
	name := filepath.Base(src)
	fileName := strings.TrimSuffix(name, filepath.Ext(name))
	sanitizedFilename, err := SanitizeFilename(fileName)
	if err != nil {
		return "", "", nil, err
	}
	dstPath := filepath.Join(outBase, sanitizedFilename)
	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return "", "", nil, err
	}

	reader, err := zip.OpenReader(src)
	if err != nil {
		return "", "", nil, err
	}
	defer reader.Close()

	var paths []string
	for _, file := range reader.File {
		path := filepath.Join(dstPath, file.Name)
		if err := ExtractFile(file, path); err != nil {
			return "", "", nil, err
		}
		paths = append(paths, path)
	}
	return fileName, dstPath, paths, nil
}

func ExtractFile(file *zip.File, dst string) error {
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
