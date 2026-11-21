package openlibrary

import (
	"context"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/platform/openlibrary"
)

// OpenLibraryCatalogLookup suggests books and authors from Open Library.
type OpenLibraryCatalogLookup struct {
	client *openlibrary.Client
}

var _ ports.ICatalogLookup = (*OpenLibraryCatalogLookup)(nil)

func NewOpenLibraryCatalogLookup(client *openlibrary.Client) *OpenLibraryCatalogLookup {
	return &OpenLibraryCatalogLookup{client: client}
}

func (o *OpenLibraryCatalogLookup) Books(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
	result, err := o.client.SearchWorks(ctx, openlibrary.SearchQuery{Query: query, Limit: limit})
	if err != nil {
		return nil, err
	}
	return bookSuggestions(result.Works), nil
}

func (o *OpenLibraryCatalogLookup) AuthorWorks(ctx context.Context, query models.AuthorWorksQuery, limit int) (models.WorksPage, error) {
	q := openlibrary.SearchQuery{
		AuthorKey: query.AuthorKey,
		Sort:      sorts[query.Sort],
		Language:  query.Language,
		Limit:     limit,
		Page:      query.Page,
	}
	if q.AuthorKey == "" {
		q.Author = query.Author
	}
	result, err := o.client.SearchWorks(ctx, q)
	if err != nil {
		return models.WorksPage{}, err
	}
	return models.WorksPage{
		Works:   bookSuggestions(result.Works),
		Total:   result.Total,
		Page:    query.Page,
		HasMore: query.Page*limit < result.Total,
	}, nil
}

// sorts maps each sort to the one of search.json; popularity is how many readers logged it.
var sorts = map[models.WorksSort]string{
	models.WorksSortPopular: "readinglog",
	models.WorksSortNew:     "new",
	models.WorksSortOld:     "old",
	models.WorksSortTitle:   "title",
}

func bookSuggestions(works []openlibrary.Work) []models.Suggestion {
	suggestions := make([]models.Suggestion, 0, len(works))
	for _, work := range works {
		suggestion := models.Suggestion{
			Kind:     models.ItemKindBook,
			Title:    strings.TrimSpace(work.Title),
			Author:   strings.Join(work.Authors, ", "),
			WorkKey:  work.Key,
			Year:     work.FirstPublishYear,
			CoverURL: openlibrary.CoverURL(work.CoverId),
		}
		if len(work.AuthorKeys) > 0 {
			suggestion.AuthorKey = work.AuthorKeys[0]
		}
		suggestions = append(suggestions, suggestion)
	}
	return suggestions
}

func (o *OpenLibraryCatalogLookup) Authors(ctx context.Context, query string, limit int) ([]models.Suggestion, error) {
	authors, err := o.client.SearchAuthors(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	suggestions := make([]models.Suggestion, 0, len(authors))
	for _, author := range authors {
		suggestions = append(suggestions, models.Suggestion{
			Kind:      models.ItemKindAuthor,
			Title:     author.Name,
			Author:    author.TopWork,
			AuthorKey: author.Key,
		})
	}
	return suggestions, nil
}
