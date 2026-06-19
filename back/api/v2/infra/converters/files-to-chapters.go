package converters

import (
	"ismelen/ermc/v2/domain"
	"ismelen/ermc/v2/infra/files"
	"mime/multipart"
	"os"
	"path/filepath"
)

func FormFilesToChapters(formFiles []*multipart.FileHeader, dstPath string) ([]*domain.Chapter, error) {
	var chapters []*domain.Chapter
	for _, formFile := range formFiles {
		switch filepath.Ext(formFile.Filename) {
		case ".zip":
			tempDir, cbzs, err := files.UnzipFormZip(formFile, dstPath)
			if err != nil {
				return nil, err
			}
			defer os.RemoveAll(tempDir)

			for _, cbz := range cbzs {
				filename, chapterPath, pages, err := files.UnzipFile(cbz, dstPath)
				if err != nil {
					return nil, err
				}
				chapters = append(chapters, domain.NewChapter(filename, chapterPath, pages, files.GetSize(cbz)))
			}
		case ".cbz":
			filename, chapterPath, pages, err := files.UnzipFormFile(formFile, dstPath)
			if err != nil {
				return nil, err
			}
			chapters = append(chapters, domain.NewChapter(filename, chapterPath, pages, formFile.Size))
		}
	}
	return chapters, nil
}
