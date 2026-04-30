package files

import (
	"archive/zip"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func UnzipFormFile(src *multipart.FileHeader, outBase string) (string, []string, error) {
	fileName := strings.TrimSuffix(src.Filename, filepath.Ext(src.Filename))
	dstPath := filepath.Join(outBase, fileName)
	if err := os.MkdirAll(dstPath, os.ModePerm); err != nil {
		return "", nil, err
	}

	file, err := src.Open()
	if err != nil { return "", nil, err }
	defer file.Close()

	reader, err := zip.NewReader(file, src.Size)
	if err != nil { return "", nil, err }
	
	var paths[]string
	for _, file := range reader.File {
		path := filepath.Join(dstPath, file.Name)
		if err := ExtractFile(file, path); err != nil {
			return "", nil, err
		}
		paths = append(paths, path)
	}
	return dstPath, paths, nil
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
