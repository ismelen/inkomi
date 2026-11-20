package openlibrary

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/platform/openlibrary"
)

const (
	matchCandidates = 5
	// minTitleMatch is the share of a candidate's title words that must appear in the query.
	minTitleMatch = 0.6
)

// OpenLibraryCatalogSource is the catalog of books: Open Library works.
type OpenLibraryCatalogSource struct {
	client *openlibrary.Client
}

var _ ports.ICatalogSource = (*OpenLibraryCatalogSource)(nil)

func NewOpenLibraryCatalogSource(client *openlibrary.Client) *OpenLibraryCatalogSource {
	return &OpenLibraryCatalogSource{client: client}
}

// MatchBook searches title (with author, or as free text that also matches authors when the
// title came from a filename like "Author - Title") and keeps the closest result.
func (o *OpenLibraryCatalogSource) MatchBook(ctx context.Context, title, author string) (*models.Work, error) {
	query := openlibrary.SearchQuery{Query: freeText(title), Limit: matchCandidates}
	if author != "" {
		query = openlibrary.SearchQuery{Title: title, Author: author, Limit: matchCandidates}
	}
	works, err := o.client.SearchWorks(ctx, query)
	if err != nil {
		return nil, err
	}

	queryWords := models.TitleWords(title + " " + author)
	var best *models.Work
	bestScore := 0.0
	for _, w := range works {
		score := models.Overlap(models.TitleWords(w.Title), queryWords)
		if author != "" && len(w.Authors) > 0 && models.Overlap(models.TitleWords(strings.Join(w.Authors, " ")), models.TitleWords(author)) == 0 {
			score /= 2
		}
		if score >= minTitleMatch && score > bestScore {
			work := toWork(w)
			best, bestScore = &work, score
		}
	}
	return best, nil
}

func (o *OpenLibraryCatalogSource) Work(ctx context.Context, key string) (*models.Work, error) {
	works, err := o.client.SearchWorks(ctx, openlibrary.SearchQuery{Query: "key:/works/" + key, Limit: 1})
	if err != nil || len(works) == 0 {
		return nil, err
	}
	work := toWork(works[0])
	return &work, nil
}

func (o *OpenLibraryCatalogSource) MatchAuthor(ctx context.Context, name string) (string, string, error) {
	authors, err := o.client.SearchAuthors(ctx, name, matchCandidates)
	if err != nil {
		return "", "", err
	}
	target := models.TitleWords(name)
	for _, author := range authors {
		if models.Overlap(target, models.TitleWords(author.Name)) >= minTitleMatch {
			return author.Key, author.Name, nil
		}
	}
	return "", "", nil
}

func (o *OpenLibraryCatalogSource) Description(ctx context.Context, workKey string) (string, error) {
	detail, err := o.client.WorkDetail(ctx, workKey)
	if errors.Is(err, openlibrary.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return detail.Description, nil
}

func (o *OpenLibraryCatalogSource) WorksByAuthor(ctx context.Context, authorKey string, limit int) ([]models.Work, error) {
	return o.search(ctx, openlibrary.SearchQuery{AuthorKey: authorKey, Limit: limit})
}

func (o *OpenLibraryCatalogSource) WorksBySubject(ctx context.Context, subject string, limit int) ([]models.Work, error) {
	works, err := o.client.SubjectWorks(ctx, subject, limit)
	if errors.Is(err, openlibrary.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return toWorks(works), nil
}

func (o *OpenLibraryCatalogSource) NewWorks(ctx context.Context, subject, authorKey string, sinceYear, limit int) ([]models.Work, error) {
	return o.search(ctx, openlibrary.SearchQuery{
		Subject:   subject,
		AuthorKey: authorKey,
		Query:     "first_publish_year:[" + strconv.Itoa(sinceYear) + " TO *]",
		Sort:      "new",
		Limit:     limit,
	})
}

func (o *OpenLibraryCatalogSource) search(ctx context.Context, query openlibrary.SearchQuery) ([]models.Work, error) {
	works, err := o.client.SearchWorks(ctx, query)
	if err != nil {
		return nil, err
	}
	return toWorks(works), nil
}

func toWorks(works []openlibrary.Work) []models.Work {
	result := make([]models.Work, len(works))
	for i, w := range works {
		result[i] = toWork(w)
	}
	return result
}

func toWork(w openlibrary.Work) models.Work {
	return models.Work{
		Key:              w.Key,
		Title:            w.Title,
		Authors:          w.Authors,
		AuthorKeys:       w.AuthorKeys,
		Subjects:         w.Subjects,
		Languages:        w.Languages,
		FirstPublishYear: w.FirstPublishYear,
		CoverId:          w.CoverId,
	}
}

// freeText drops the characters the search engine reads as operators: in "Author - Title" the
// dash would exclude the title.
func freeText(value string) string {
	return strings.Join(strings.FieldsFunc(value, func(r rune) bool {
		return strings.ContainsRune(`-+!():^[]"{}~*?\/&|`, r) || r == ' '
	}), " ")
}
