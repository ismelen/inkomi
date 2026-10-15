package searchbooks

import (
	"context"
	"strings"

	"github.com/ismelen/inkomi/back/library-service/internal/domain/bookfilter"
	"github.com/ismelen/inkomi/back/library-service/internal/domain/models"
	"github.com/ismelen/inkomi/back/library-service/internal/domain/ports"
	"github.com/ismelen/inkomi/back/library-service/internal/shared/filter"
)

type SearchBooksUC struct {
	library ports.ILibrary
}

func NewSearchBooksUC(library ports.ILibrary) *SearchBooksUC {
	return &SearchBooksUC{library: library}
}

// Execute searches query and keeps the deduplicated books in language and formats; empty
// filters keep every book.
func (s *SearchBooksUC) Execute(ctx context.Context, query, language string, formats []string) ([]models.Book, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, models.ErrEmptyQuery
	}

	books, err := s.library.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	chain := filter.Use(
		&bookfilter.LanguageFilter{Language: language},
		&bookfilter.FormatFilter{Formats: formats},
		&bookfilter.DeduplicateFilter{},
	)
	_, books = chain.Filter(books)
	return books, nil
}
