package expandcatalog

import (
	"context"
	"log"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

const (
	topAuthors      = 15
	topSubjects     = 10
	worksPerAuthor  = 20
	worksPerSubject = 40
	newPerQuery     = 20
	// newYears is how far back a work still counts as new.
	newYears = 1
)

// ExpandCatalogUC adds candidates to the catalog: other works of the users' authors, works of
// their most common subjects and what those authors and subjects published lately.
type ExpandCatalogUC struct {
	source  ports.ICatalogSource
	catalog ports.ICatalogRepository
	now     func() time.Time
}

func NewExpandCatalogUC(source ports.ICatalogSource, catalog ports.ICatalogRepository, now func() time.Time) *ExpandCatalogUC {
	return &ExpandCatalogUC{source: source, catalog: catalog, now: now}
}

// Execute fetches candidates while budget lasts and returns how many works were new.
func (e *ExpandCatalogUC) Execute(ctx context.Context, budget *models.Budget) (int, error) {
	authors, err := e.catalog.TopAuthorKeys(ctx, topAuthors)
	if err != nil {
		return 0, err
	}
	subjects, err := e.catalog.TopSubjects(ctx, topSubjects)
	if err != nil {
		return 0, err
	}
	sinceYear := e.now().Year() - newYears

	// novelties first: they change the most between runs
	var queries []func() ([]models.Work, error)
	for _, author := range authors {
		queries = append(queries, func() ([]models.Work, error) {
			return e.source.NewWorks(ctx, "", author, sinceYear, newPerQuery)
		})
	}
	for _, subject := range subjects {
		queries = append(queries, func() ([]models.Work, error) {
			return e.source.NewWorks(ctx, subject, "", sinceYear, newPerQuery)
		})
	}
	for _, author := range authors {
		queries = append(queries, func() ([]models.Work, error) {
			return e.source.WorksByAuthor(ctx, author, worksPerAuthor)
		})
	}
	for _, subject := range subjects {
		queries = append(queries, func() ([]models.Work, error) {
			return e.source.WorksBySubject(ctx, subject, worksPerSubject)
		})
	}

	added := 0
	for _, query := range queries {
		if ctx.Err() != nil {
			return added, ctx.Err()
		}
		if !budget.Take(1) {
			break
		}
		works, err := query()
		if err != nil {
			log.Printf("recommendations: expand catalog: %v", err)
			continue
		}
		for i := range works {
			works[i].Origin = models.OriginCandidate
			created, err := e.catalog.UpsertWork(ctx, &works[i])
			if err != nil {
				return added, err
			}
			if created {
				added++
			}
		}
	}
	return added, nil
}
