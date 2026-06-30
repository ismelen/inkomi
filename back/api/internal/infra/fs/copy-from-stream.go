package fs

import (
	"io"
	"os"
	"path/filepath"
)

func CopyFromStream(stream io.ReadCloser, dstPath string) (string, error) {
	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return "", nil
	}
	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, stream)
	return dstPath, err
}
