package backfillworks

import (
	"context"
	"log"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/resolvework"
)

// BackfillWorksUC resolves the wished books still without catalog details: those added
// before details existed, and those whose lookup failed when they were added.
type BackfillWorksUC struct {
	items    ports.IItemRepository
	resolver *resolvework.ResolveWorkUC
}

func NewBackfillWorksUC(items ports.IItemRepository, resolver *resolvework.ResolveWorkUC) *BackfillWorksUC {
	return &BackfillWorksUC{items: items, resolver: resolver}
}

// Start runs Execute right away and then every interval until ctx is done.
func (b *BackfillWorksUC) Start(ctx context.Context, interval time.Duration) {
	go func() {
		b.run(ctx)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				b.run(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (b *BackfillWorksUC) run(ctx context.Context) {
	resolved, failed, err := b.Execute(ctx)
	if err != nil {
		log.Printf("wishlist: backfill works: %v", err)
		return
	}
	if resolved > 0 || failed > 0 {
		log.Printf("wishlist: backfill works: %d resolved, %d failed", resolved, failed)
	}
}

// Execute resolves every pending book and returns how many got a cover and how many failed;
// the failed ones are tried again on the next run.
func (b *BackfillWorksUC) Execute(ctx context.Context) (resolved, failed int, err error) {
	items, err := b.items.ListBooksWithoutWork(ctx)
	if err != nil {
		return 0, 0, err
	}
	for i := range items {
		if ctx.Err() != nil {
			return resolved, failed, ctx.Err()
		}
		if err := b.resolver.Execute(ctx, &items[i], ""); err != nil {
			failed++
			continue
		}
		if items[i].CoverURL != "" {
			resolved++
		}
	}
	return resolved, failed, nil
}
