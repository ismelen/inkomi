package bookfilter

import (
	"regexp"
	"strings"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/shared/filter"
	"github.com/ismelen/inkomi/back/library-service/internal/shared/strutil"
)

var (
	isbnListRe = regexp.MustCompile(`^[\d\s;\-\.,xX]+$`)
	spamWords  = []string{
		"downloaden", "gratis", "descargar gratis", "pdf gratis",
		"epub gratis", "lesen", "télécharger",
	}
)

// DeduplicateFilter drops spam titles and keeps one book per normalized title, preferring
// one with an md5. It ends the chain.
type DeduplicateFilter struct {
	filter.Base[[]models.Book, []models.Book]
}

func (f *DeduplicateFilter) Filter(books []models.Book) (bool, []models.Book) {
	order := []string{}
	best := map[string]models.Book{}

	for _, b := range books {
		if len([]rune(b.Title)) < 3 || isSpamTitle(b.Title) {
			continue
		}

		key := strutil.Normalize(b.Title)
		current, exists := best[key]
		if !exists {
			order = append(order, key)
		}
		if !exists || (current.MD5 == "" && b.MD5 != "") {
			best[key] = b
		}
	}

	out := make([]models.Book, 0, len(order))
	for _, key := range order {
		out = append(out, best[key])
	}
	return true, out
}

func isSpamTitle(title string) bool {
	if len(title) > 6 && isbnListRe.MatchString(title) {
		return true
	}

	lower := strings.ToLower(title)
	for _, w := range spamWords {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}
