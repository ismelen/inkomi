package bookfilter

import (
	"slices"
	"strings"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/shared/filter"
)

// FormatFilter keeps the books whose extension is in Formats; no Formats keeps them all.
type FormatFilter struct {
	filter.Base[[]models.Book, []models.Book]
	Formats []string
}

func (f *FormatFilter) Filter(books []models.Book) (bool, []models.Book) {
	if len(f.Formats) == 0 {
		return f.Next(books)
	}

	filtered := make([]models.Book, 0, len(books))
	for _, b := range books {
		if slices.ContainsFunc(f.Formats, func(format string) bool { return strings.EqualFold(b.Extension, format) }) {
			filtered = append(filtered, b)
		}
	}
	return f.Next(filtered)
}
