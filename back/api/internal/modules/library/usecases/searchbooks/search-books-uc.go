package searchbooks

import (
	"context"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/bookfilter"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/shared/filter"
)

const (
	// minBooks is how many filtered books a search tries to return before stopping.
	minBooks = 10
	// maxRounds bounds the library pages read for one search.
	maxRounds = 2
)

type SearchBooksUC struct {
	library ports.ILibrary
}

func NewSearchBooksUC(library ports.ILibrary) *SearchBooksUC {
	return &SearchBooksUC{library: library}
}

// Execute searches query in field from page on and keeps the deduplicated books in language and
// formats; empty filters keep every book. When the filters leave few books it reads the next
// pages too, up to maxRounds, so a search in a minority language doesn't come back empty.
func (s *SearchBooksUC) Execute(ctx context.Context, query string, field models.SearchField, language string, formats []string, page int) (models.SearchPage, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return models.SearchPage{}, models.ErrEmptyQuery
	}
	if !field.IsValid() {
		return models.SearchPage{}, models.ErrInvalidField
	}
	if page < 1 {
		return models.SearchPage{}, models.ErrInvalidPage
	}

	chain := filter.Use(
		&bookfilter.LanguageFilter{Language: language},
		&bookfilter.FormatFilter{Formats: formats},
		&bookfilter.DeduplicateFilter{},
	)

	var raw, books []models.Book
	next := page
	for round := 0; round < maxRounds && next > 0; round++ {
		result, err := s.library.Search(ctx, query, field, next)
		if err != nil {
			if round > 0 {
				break
			}
			return models.SearchPage{}, err
		}
		raw = append(raw, result.Books...)
		_, books = chain.Filter(raw)
		next = result.NextPage
		if len(books) >= minBooks {
			break
		}
	}
	if books == nil {
		books = []models.Book{}
	}
	return models.SearchPage{Books: books, NextPage: next}, nil
}
