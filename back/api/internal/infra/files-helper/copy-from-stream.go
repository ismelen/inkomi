package filesHelper

import (
	"io"
	"os"
)

func CopyFromStream(stream io.ReadCloser, dstPath string) (string, error) {
	out, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, stream)
	return dstPath, err
}
