package refreshcatalog

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/computerecommendations"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/embedworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/expandcatalog"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/importsignals"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/resolvesignals"
)

const (
	// candidateTTL is how long a candidate nobody's taste brings back stays in the catalog.
	candidateTTL = 90 * 24 * time.Hour
	// quickBudget is the request budget of the pass that follows a wishlist change.
	quickBudget = 30
	// debounce groups the wishlist changes made in a row into one quick pass.
	debounce = 20 * time.Second
	// startDelay leaves the startup alone before the first pass.
	startDelay = time.Minute
)

// RefreshCatalogUC keeps the recommender up to date without anyone doing it by hand: every
// interval it imports the signals, matches them, grows the catalog, embeds what is new, prunes
// the stale candidates and recomputes the recommendations. A wishlist change triggers a quick
// pass (no catalog growth) so the user sees its effect soon.
type RefreshCatalogUC struct {
	importUC  *importsignals.ImportSignalsUC
	resolveUC *resolvesignals.ResolveSignalsUC
	expandUC  *expandcatalog.ExpandCatalogUC
	embedUC   *embedworks.EmbedWorksUC
	computeUC *computerecommendations.ComputeRecommendationsUC
	catalog   ports.ICatalogRepository
	signals   ports.ISignalRepository
	budget    int
	now       func() time.Time

	mu      sync.Mutex
	trigger chan struct{}
}

func NewRefreshCatalogUC(
	importUC *importsignals.ImportSignalsUC,
	resolveUC *resolvesignals.ResolveSignalsUC,
	expandUC *expandcatalog.ExpandCatalogUC,
	embedUC *embedworks.EmbedWorksUC,
	computeUC *computerecommendations.ComputeRecommendationsUC,
	catalog ports.ICatalogRepository,
	signals ports.ISignalRepository,
	budget int,
	now func() time.Time,
) *RefreshCatalogUC {
	return &RefreshCatalogUC{
		importUC:  importUC,
		resolveUC: resolveUC,
		expandUC:  expandUC,
		embedUC:   embedUC,
		computeUC: computeUC,
		catalog:   catalog,
		signals:   signals,
		budget:    budget,
		now:       now,
		trigger:   make(chan struct{}, 1),
	}
}

// Start runs a full pass right away and then every interval, and the quick passes asked by
// Trigger, until ctx is done.
func (r *RefreshCatalogUC) Start(ctx context.Context, interval time.Duration) {
	go func() {
		select {
		case <-time.After(startDelay):
		case <-ctx.Done():
			return
		}
		r.run(ctx, true)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				r.run(ctx, true)
			case <-r.trigger:
				select {
				case <-time.After(debounce):
				case <-ctx.Done():
					return
				}
				r.run(ctx, false)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Trigger asks for a quick pass; calls made while one is pending are merged.
func (r *RefreshCatalogUC) Trigger(int) {
	select {
	case r.trigger <- struct{}{}:
	default:
	}
}

func (r *RefreshCatalogUC) run(ctx context.Context, full bool) {
	if err := r.Execute(ctx, full); err != nil && ctx.Err() == nil {
		log.Printf("recommendations: refresh: %v", err)
	}
}

// Execute runs one pass; a full one also grows and prunes the catalog and recomputes everyone.
func (r *RefreshCatalogUC) Execute(ctx context.Context, full bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	started := r.now()

	if err := r.importUC.Execute(ctx); err != nil {
		return err
	}
	budget := models.NewBudget(quickBudget)
	if full {
		budget = models.NewBudget(r.budget)
	}
	matched, err := r.resolveUC.Execute(ctx, budget)
	if err != nil {
		return err
	}

	added, pruned := 0, 0
	if full {
		if added, err = r.expandUC.Execute(ctx, budget); err != nil {
			return err
		}
		if pruned, err = r.catalog.PruneCandidates(ctx, r.now().Add(-candidateTTL)); err != nil {
			return err
		}
	}
	embedded, err := r.embedUC.Execute(ctx)
	if err != nil {
		// without embeddings the ranking still works with the ones already stored
		log.Printf("recommendations: embed works: %v", err)
	}

	users, err := r.signals.TakeStale(ctx)
	if err != nil {
		return err
	}
	if full && (embedded > 0 || pruned > 0) {
		// the catalog changed for everyone
		if users, err = r.signals.ListUsers(ctx); err != nil {
			return err
		}
	}
	if err := r.computeUC.Execute(ctx, users); err != nil {
		return err
	}
	log.Printf("recommendations: %s pass in %s: %d matched, %d new works, %d pruned, %d embedded, %d users ranked, %d requests left",
		passName(full), r.now().Sub(started).Round(time.Second), matched, added, pruned, embedded, len(users), budget.Left())
	return nil
}

func passName(full bool) string {
	if full {
		return "full"
	}
	return "quick"
}
