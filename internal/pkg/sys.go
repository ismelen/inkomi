package pkg

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func NewTempDir(suffix string) (string, error) {
	tmp, err := os.MkdirTemp("", "ERMC-"+suffix)
	if err != nil {
		return "", err
	}

	return tmp, nil
}

func CopyDirFiles(srcDir, dstDir string) error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if IsImage(entry.Name()) {
			if err := CopyFile(
				filepath.Join(srcDir, entry.Name()),
				filepath.Join(dstDir, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	return nil
}

func GetChildsInfo(path string) (infos []Pair[string, int64], err error) {
	err = filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		infos = append(infos, Pair[string, int64]{
			Fst: path,
			Snd: info.Size(),
		})

		return nil
	})

	return
}
