package bookfilter

import (
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/shared/filter"
)

// LanguageFilter keeps the books in Language, also those that list it among several
// ("Spanish; English"); an empty Language keeps them all.
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
		if hasLanguage(b.Language, f.Language) {
			filtered = append(filtered, b)
		}
	}
	return f.Next(filtered)
}

func hasLanguage(languages, language string) bool {
	for _, l := range strings.FieldsFunc(languages, func(r rune) bool { return r == ',' || r == ';' || r == '/' }) {
		if strings.EqualFold(strings.TrimSpace(l), language) {
			return true
		}
	}
	return false
}
