package getwork

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
)

const (
	// catalogLanguage is the language the catalog descriptions are written in
	catalogLanguage = "en"
	// missingSynopsisTTL is how long a synopsis that wasn't found is not asked for again
	missingSynopsisTTL = 30 * 24 * time.Hour
)

// GetWorkUC returns the details of a work (synopsis, cover...), from the stored ones or else
// from the catalog, which it then stores. With a language, the synopsis is the one of an edition
// in it, or none.
type GetWorkUC struct {
	works    ports.IWorkRepository
	catalog  ports.ICatalogLookup
	synopses ports.ISynopsisRepository
	lookup   ports.ISynopsisLookup
}

func NewGetWorkUC(works ports.IWorkRepository, catalog ports.ICatalogLookup, synopses ports.ISynopsisRepository, synopsisLookup ports.ISynopsisLookup) *GetWorkUC {
	return &GetWorkUC{works: works, catalog: catalog, synopses: synopses, lookup: synopsisLookup}
}

// Execute takes lang as an ISO 639-1 code; "" keeps the catalog description.
func (g *GetWorkUC) Execute(ctx context.Context, workKey, lang string) (*models.WorkDetails, error) {
	if !models.IsWorkKey(workKey) {
		return nil, models.ErrInvalidWorkKey
	}
	if lang != "" && !models.IsLanguage(lang) {
		return nil, models.ErrInvalidLanguage
	}
	work, err := g.details(ctx, workKey)
	if err != nil || lang == "" {
		return work, err
	}

	synopsis := g.synopsis(ctx, workKey, lang)
	if synopsis != "" || lang != catalogLanguage {
		work.Description = synopsis
	}
	return work, nil
}

func (g *GetWorkUC) details(ctx context.Context, workKey string) (*models.WorkDetails, error) {
	stored, err := g.works.Get(ctx, workKey)
	if err == nil {
		return stored, nil
	}
	if !errors.Is(err, models.ErrWorkNotFound) {
		return nil, err
	}

	work, err := g.catalog.Work(ctx, workKey)
	if errors.Is(err, models.ErrWorkNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %w", models.ErrLookupFailure, err)
	}
	work.WorkKey = workKey
	if err := g.works.Save(ctx, work); err != nil {
		// the details are still good to answer with; the next request fetches them again
		log.Printf("wishlist: save work %s: %v", workKey, err)
	}
	return &work, nil
}

// synopsis returns the stored synopsis or else looks it up and stores it. A failed lookup
// answers with none and isn't stored, so the next request retries it.
func (g *GetWorkUC) synopsis(ctx context.Context, workKey, lang string) string {
	stored, err := g.synopses.Get(ctx, workKey, lang)
	switch {
	case err == nil && (stored.Text != "" || time.Since(stored.FetchedAt) < missingSynopsisTTL):
		return stored.Text
	case err != nil && !errors.Is(err, models.ErrSynopsisNotFound):
		log.Printf("wishlist: get synopsis %s/%s: %v", workKey, lang, err)
		return ""
	}

	text, err := g.lookup.Synopsis(ctx, workKey, lang)
	if err != nil {
		log.Printf("wishlist: look up synopsis %s/%s: %v", workKey, lang, err)
		return ""
	}
	if err := g.synopses.Save(ctx, models.Synopsis{WorkKey: workKey, Language: lang, Text: text}); err != nil {
		log.Printf("wishlist: save synopsis %s/%s: %v", workKey, lang, err)
	}
	return text
}
