package filters

import (
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/shared/filter"
	"strings"
)

type FormatFilter struct {
	filter.Base[[]book.Book, []book.Book]
	Formats []string
}

func (f *FormatFilter) Filter(books []book.Book) (bool, []book.Book) {
	if len(f.Formats) == 0 {
		return f.Next(books)
	}

	var filtered []book.Book
	for _, b := range books {
		for _, format := range f.Formats {
			if strings.EqualFold(b.Extension, format) {
				filtered = append(filtered, b)
				break
			}
		}
	}
	return f.Next(filtered)
}
