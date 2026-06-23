package libgen

import (
	"ismelen/inkomi/internal/domain"
	"ismelen/inkomi/internal/infra/helpers"
	"ismelen/inkomi/internal/ports"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
)

type LibgenService struct {
	mirror atomic.Value
}

func New() *LibgenService {
	return &LibgenService{}
}

func (l *LibgenService) Search(query string, language string, formats []string) ([]domain.Book, error) {
	mirror := l.mirror.Load().(ports.LibgenMirror)
	books, err := mirror.Search(query)
	if err != nil {
		return nil, err
	}

	books = filterBooks(books, language, formats)
	return deduplicateBooks(books), nil
}

func (l *LibgenService) Download(request domain.LibgenDownloadRequestDTO) (*domain.LibgenDownloadResultDTO, error) {
	mirror := l.mirror.Load().(ports.LibgenMirror)
	return mirror.Download(request)
}

func filterBooks(books []domain.Book, language string, formats []string) []domain.Book {
	if language == "" && len(formats) == 0 {
		return books
	}

	filters := [](func(book domain.Book) bool){}

	if language != "" {
		filters = append(filters, func(book domain.Book) bool {
			return strings.EqualFold(book.Language, language)
		})
	}

	if len(formats) > 0 {
		filters = append(filters, func(book domain.Book) bool {
			for _, s := range formats {
				if strings.EqualFold(s, book.Extension) {
					return true
				}
			}
			return false
		})
	}

	var out []domain.Book

bookLoop:
	for _, book := range books {
		for _, filter := range filters {
			if filter(book) {
				continue bookLoop
			}
		}
		out = append(out, book)
	}

	return out
}

func deduplicateBooks(books []domain.Book) []domain.Book {
	type group struct {
		best      domain.Book
		bestPages int
		count     int
	}
	order := []string{}
	groups := map[string]*group{}

	for _, b := range books {
		if isSpamTitle(b.Title) {
			continue
		}
		if len([]rune(b.Title)) < 3 {
			continue
		}
		key := helpers.NormalizeString(b.Title)
		if key == "" {
			key = "__" + b.ID
		}
		pages := pageCount(b)
		if g, exists := groups[key]; exists {
			g.count++
			if pages > g.bestPages || (g.best.MD5 == "" && b.MD5 != "") {
				g.best = b
				g.bestPages = pages
			}
		} else {
			groups[key] = &group{best: b, bestPages: pages, count: 1}
			order = append(order, key)
		}
	}

	out := make([]domain.Book, 0, len(order))
	for _, key := range order {
		out = append(out, groups[key].best)
	}
	return out
}

func pageCount(b domain.Book) int {
	n, _ := strconv.Atoi(strings.TrimSpace(b.Pages))
	return n
}

var isbnSpamRe = regexp.MustCompile(`^[\d\s;\-\.,xX]+$`)

func isSpamTitle(title string) bool {
	if len(title) > 6 {
		isIsbnList := isbnSpamRe.MatchString(title)
		if isIsbnList {
			return true
		}
	}
	t := strings.ToLower(title)
	spamWords := []string{
		"downloaden", "gratis", "descargar gratis", "pdf gratis",
		"epub gratis", "lesen", "télécharger",
	}
	for _, w := range spamWords {
		if strings.Contains(t, w) {
			return true
		}
	}
	return false
}
