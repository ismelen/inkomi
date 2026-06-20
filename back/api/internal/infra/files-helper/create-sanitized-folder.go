package filesHelper

import (
	"os"
	"path/filepath"
	"strings"
)

func CreateSanitizedFolder(name string, outBase string) (string, error) {
	basePath := outBase
	if name != "" {
		srcName := strings.TrimSuffix(name, filepath.Ext(name))
		sanitizedSrcName, err := SanitizeFilename(srcName)
		if err != nil {
			return "", err
		}
		basePath = filepath.Join(outBase, sanitizedSrcName)
	}

	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		return "", err
	}

	return basePath, nil
}
