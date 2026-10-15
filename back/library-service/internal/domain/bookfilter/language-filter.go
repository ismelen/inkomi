package bookfilter

import (
	"strings"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/shared/filter"
)

// LanguageFilter keeps the books in Language; an empty Language keeps them all.
type LanguageFilter struct {
	filter.Base[[]models.Book, []models.Book]
	Language string
}

func (f *LanguageFilter) Filter(books []models.Book) (bool, []models.Book) {
	if f.Language == "" {
		return f.Next(books)
	}

	filtered := make([]models.Book, 0, len(books))
	for _, b := range books {
		if strings.EqualFold(b.Language, f.Language) {
			filtered = append(filtered, b)
		}
	}
	return f.Next(filtered)
}
