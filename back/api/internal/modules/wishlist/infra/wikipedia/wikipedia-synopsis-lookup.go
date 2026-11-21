package wikipedia

import (
	"context"
	"log"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/ports"
	"github.com/ismelen/inkomi/back/api/internal/platform/openlibrary"
	"github.com/ismelen/inkomi/back/api/internal/platform/wikimedia"
)

// WikipediaSynopsisLookup uses the lead of the Wikipedia article in the language about the work,
// found on Wikidata by the work key or by the key of its edition in that language.
type WikipediaSynopsisLookup struct {
	editions *openlibrary.Client
	wiki     *wikimedia.Client
}

var _ ports.ISynopsisLookup = (*WikipediaSynopsisLookup)(nil)

func NewWikipediaSynopsisLookup(editions *openlibrary.Client, wiki *wikimedia.Client) *WikipediaSynopsisLookup {
	return &WikipediaSynopsisLookup{editions: editions, wiki: wiki}
}

func (w *WikipediaSynopsisLookup) Synopsis(ctx context.Context, workKey, lang string) (string, error) {
	keys := []string{workKey}
	editionKey, err := w.editions.EditionKeyIn(ctx, workKey, lang)
	if err != nil {
		// the work key alone usually finds the article
		log.Printf("wishlist: edition of %s in %s: %v", workKey, lang, err)
	}
	if editionKey != "" {
		keys = append(keys, editionKey)
	}

	title, err := w.wiki.ArticleFor(ctx, keys, lang)
	if err != nil || title == "" {
		return "", err
	}
	return w.wiki.Summary(ctx, lang, title)
}
