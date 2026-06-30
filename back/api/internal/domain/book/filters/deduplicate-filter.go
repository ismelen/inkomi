package filters

import (
	"ismelen/inkomi/internal/domain/book"
	"ismelen/inkomi/internal/shared/filter"
	"ismelen/inkomi/internal/shared/strutil"
	"regexp"
	"strconv"
	"strings"
)

type DeduplicateFilter struct {
	filter.Base[[]book.Book, []book.Book]
}

func (f *DeduplicateFilter) Filter(books []book.Book) (bool, []book.Book) {
	type group struct {
		best      book.Book
		bestPages int
		count     int
	}
	var order []string
	groups := map[string]*group{}

	for _, b := range books {
		if isSpamTitle(b.Title) {
			continue
		}
		if len([]rune(b.Title)) < 3 {
			continue
		}
		key := strutil.NormalizeString(b.Title)
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

	var out []book.Book
	for _, key := range order {
		out = append(out, groups[key].best)
	}
	return true, out
}

func pageCount(b book.Book) int {
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
