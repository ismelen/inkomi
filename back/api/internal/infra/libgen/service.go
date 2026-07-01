package libgen

import (
	"ismelen/inkomi/internal/domain/book"
	bookfilters "ismelen/inkomi/internal/domain/book/filters"
	"ismelen/inkomi/internal/shared/filter"
	"sync/atomic"
)

type LibgenService struct {
	mirror atomic.Value
}

func New() *LibgenService {
	return &LibgenService{}
}

func (l *LibgenService) Search(query string, language string, formats []string) ([]book.Book, error) {
	mirror := l.mirror.Load().(book.LibgenMirror)
	books, err := mirror.Search(query)
	if err != nil {
		return nil, err
	}

	filterChain := filter.Use(
		&bookfilters.LanguageFilter{Language: language},
		&bookfilters.FormatFilter{Formats: formats},
		&bookfilters.DeduplicateFilter{},
	)

	_, filteredBooks := filterChain.Filter(books)
	return filteredBooks, nil
}

func (l *LibgenService) Download(md5 string) (*book.LibgenDownload, error) {
	mirror := l.mirror.Load().(book.LibgenMirror)
	return mirror.Download(md5)
}
