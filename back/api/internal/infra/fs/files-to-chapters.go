package fs

import (
	"ismelen/inkomi/internal/domain/manga"
)

func FileToChapter(file string, destPath string) (*manga.Chapter, error) {
	filename, chapterPath, pages, err := UnzipFile(file, destPath)
	if err != nil {
		return nil, err
	}
	return manga.NewChapter(filename, chapterPath, pages), nil
}
