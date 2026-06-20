package converters

import (
	"ismelen/inkomi/internal/domain"
	filesHelper "ismelen/inkomi/internal/infra/files-helper"
)

func FileToChapter(file string, destPath string) (*domain.Chapter, error) {
	filename, chapterPath, pages, err := filesHelper.UnzipFile(file, destPath)
	if err != nil {
		return nil, err
	}
	return domain.NewChapter(filename, chapterPath, pages, filesHelper.GetSize(file)), nil
}
