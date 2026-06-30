package filters

import (
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/shared/filter"
	"strings"
)

type LanguageFilter struct {
	filter.Base[[]book.Book, []book.Book]
	Language string
}

func (f *LanguageFilter) Filter(books []book.Book) (bool, []book.Book) {
	if f.Language == "" {
		return f.Next(books)
	}

	var filtered []book.Book
	for _, b := range books {
		if strings.EqualFold(b.Language, f.Language) {
			filtered = append(filtered, b)
		}
	}
	return f.Next(filtered)
}
