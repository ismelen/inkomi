package converters

import (
	"ismelen/ermc/v2/domain"
	"ismelen/ermc/v2/infra/files"
	"mime/multipart"
)

func FormFilesToChapters(formFiles []*multipart.FileHeader, dstPath string) ([]*domain.Chapter, error) {
	var chapters []*domain.Chapter
	for _, formFile := range formFiles {
		chapterPath, pages, err := files.UnzipFormFile(formFile, dstPath)
		if err != nil {
			return nil, err
		}
		chapters = append(chapters, domain.NewChapter(chapterPath, pages))
	}
	return chapters, nil
}