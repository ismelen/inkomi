package pkg

import (
	"archive/zip"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func UnizpFile(src, dir string) ([]string, error) {
	filename := filepath.Base(src)
	ext := filepath.Ext(filename)
	dirName := strings.TrimSuffix(filename, ext)
	dstPath := filepath.Join(dir, dirName)
	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return nil, err
	}

	reader, err := zip.OpenReader(src)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, file := range reader.File {
		path := filepath.Join(dstPath, file.Name)
		if err = ExtractFile(file, path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}

	return paths, nil
}

func UnzipFileForm(src *multipart.FileHeader, dir string) ([]string, error) {
	ext := filepath.Ext(src.Filename)
	dirName := strings.TrimSuffix(src.Filename, ext)
	dstPath := filepath.Join(dir, dirName)
	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return nil, err
	}

	file, err := src.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader, err := zip.NewReader(file, src.Size)
	if err != nil {
		return nil, err
	}

	var paths []string
	for _, file := range reader.File {
		path := filepath.Join(dstPath, file.Name)
		if err := ExtractFile(file, path); err != nil {
			return nil, err
		}
		paths = append(paths, path)
	}

	return paths, nil
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
