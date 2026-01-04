package ZipUtils

import (
	"archive/zip"
	"io"
	FileUtils "ismelen/ermc/internal/utils/file"
	"os"
	"path/filepath"
	"slices"
)

func UnzipFile(src, dst, dstFileName string) (newDstPath string, extractedFilePaths []string, err error) {
	newDstPath = filepath.Join(dst, dstFileName)

	if err = os.MkdirAll(newDstPath, os.ModePerm); err != nil {
		return
	}

	r, err := zip.OpenReader(src)
	if err != nil {
		return
	}
	defer r.Close()

	for _, file := range r.File {
		if file.FileInfo().IsDir() || !FileUtils.IsImage(file.Name) {
			continue
		}

		dstFileName := filepath.Join(newDstPath, file.Name)
		if err = extractFile(file, dstFileName); err != nil {
			return
		}
		extractedFilePaths = append(extractedFilePaths, dstFileName)
	}

	slices.SortFunc(extractedFilePaths, FileUtils.FilenameCmp)

	return
}

func extractFile(f *zip.File, dest string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}

	_, err = io.Copy(out, rc)

	rc.Close()
	out.Close()
	return err
}
