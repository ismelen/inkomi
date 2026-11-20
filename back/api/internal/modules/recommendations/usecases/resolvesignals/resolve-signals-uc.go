package resolvesignals

import (
	"context"
	"errors"
	"log"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

const (
	maxAttempts = 3
	batch       = 100
)

// ResolveSignalsUC matches the pending signals against the catalog: a book becomes a user work
// of the catalog (with its description) and an author gets its key.
type ResolveSignalsUC struct {
	source  ports.ICatalogSource
	catalog ports.ICatalogRepository
	signals ports.ISignalRepository
}

func NewResolveSignalsUC(source ports.ICatalogSource, catalog ports.ICatalogRepository, signals ports.ISignalRepository) *ResolveSignalsUC {
	return &ResolveSignalsUC{source: source, catalog: catalog, signals: signals}
}

// Execute resolves pending signals while budget lasts and returns how many it matched. A failed
// request leaves the signal for the next run; after maxAttempts it is given up.
func (r *ResolveSignalsUC) Execute(ctx context.Context, budget *models.Budget) (int, error) {
	pending, err := r.signals.ListPending(ctx, batch, maxAttempts)
	if err != nil {
		return 0, err
	}

	matched := 0
	for _, signal := range pending {
		if ctx.Err() != nil {
			return matched, ctx.Err()
		}
		ok, err := r.resolve(ctx, budget, signal)
		if errors.Is(err, errNoBudget) {
			break
		}
		if err != nil {
			log.Printf("recommendations: match %q of user %d: %v", signal.Title, signal.UserId, err)
			if err := r.signals.Resolve(ctx, signal.Id, models.SignalPending, ""); err != nil {
				return matched, err
			}
			continue
		}
		if ok {
			matched++
		}
	}
	return matched, nil
}

var errNoBudget = errors.New("no request budget left")

func (r *ResolveSignalsUC) resolve(ctx context.Context, budget *models.Budget, signal models.Signal) (bool, error) {
	if signal.Kind == models.SignalAuthor {
		return r.resolveAuthor(ctx, budget, signal)
	}
	return r.resolveBook(ctx, budget, signal)
}

func (r *ResolveSignalsUC) resolveAuthor(ctx context.Context, budget *models.Budget, signal models.Signal) (bool, error) {
	key := signal.MatchedKey
	if key == "" {
		if !budget.Take(1) {
			return false, errNoBudget
		}
		var err error
		if key, _, err = r.source.MatchAuthor(ctx, signal.Title); err != nil {
			return false, err
		}
	}
	return r.finish(ctx, signal, key)
}

func (r *ResolveSignalsUC) resolveBook(ctx context.Context, budget *models.Budget, signal models.Signal) (bool, error) {
	// the work, its description
	if !budget.Take(2) {
		return false, errNoBudget
	}
	var work *models.Work
	var err error
	if signal.MatchedKey != "" {
		work, err = r.source.Work(ctx, signal.MatchedKey)
	} else {
		work, err = r.source.MatchBook(ctx, signal.Title, signal.Author)
	}
	if err != nil {
		return false, err
	}
	if work == nil {
		return r.finish(ctx, signal, "")
	}

	if description, err := r.source.Description(ctx, work.Key); err != nil {
		log.Printf("recommendations: description of %s: %v", work.Key, err)
	} else {
		work.Description = description
	}
	work.Origin = models.OriginUser
	if _, err := r.catalog.UpsertWork(ctx, work); err != nil {
		return false, err
	}
	return r.finish(ctx, signal, work.Key)
}

// finish stores the match (or the miss, when key is empty) and marks the user for a recompute.
func (r *ResolveSignalsUC) finish(ctx context.Context, signal models.Signal, key string) (bool, error) {
	if key == "" {
		status := models.SignalPending
		if signal.Attempts+1 >= maxAttempts {
			status = models.SignalUnmatched
		}
		return false, r.signals.Resolve(ctx, signal.Id, status, "")
	}
	if err := r.signals.Resolve(ctx, signal.Id, models.SignalMatched, key); err != nil {
		return false, err
	}
	return true, r.signals.MarkStale(ctx, signal.UserId)
}
