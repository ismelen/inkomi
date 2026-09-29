package usecases

import (
	"context"
	"fmt"
	"ismelen/inkomi/internal/domain/book"
	booksFilter "ismelen/inkomi/internal/domain/book/filters"
	"ismelen/inkomi/internal/shared/filter"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type SearchBookUC struct {
	provider book.BooksProvider
}

func NewSearchBookUC(provider book.BooksProvider) *SearchBookUC {
	return &SearchBookUC{
		provider: provider,
	}
}

func (s *SearchBookUC) Execute(ctx context.Context, query string, language string, formats []string) ([]book.Book, error) {
	ctx, span := otel.Tracer("inkomi-api").Start(ctx, "SearchBookUC.Execute")
	defer span.End()

	mirror, ok := s.provider.GetMirror()
	if !ok {
		err := fmt.Errorf("no mirror available yet")
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	books, err := mirror.Search(query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	filterChain := filter.Use(
		&booksFilter.LanguageFilter{Language: language},
		&booksFilter.FormatFilter{Formats: formats},
		&booksFilter.DeduplicateFilter{},
	)

	_, filteredBooks := filterChain.Filter(books)
	return filteredBooks, nil
}
