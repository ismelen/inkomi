package importsignals

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

const (
	sentCursor = "sent_cursor"
	pageSize   = 500
)

// ImportSignalsUC brings what users sent (since the last import) and what they wish into the
// signals of the recommender, and drops the wishes they removed.
type ImportSignalsUC struct {
	sent     ports.ISentBooks
	wishlist ports.IWishlist
	signals  ports.ISignalRepository
}

func NewImportSignalsUC(sent ports.ISentBooks, wishlist ports.IWishlist, signals ports.ISignalRepository) *ImportSignalsUC {
	return &ImportSignalsUC{sent: sent, wishlist: wishlist, signals: signals}
}

func (i *ImportSignalsUC) Execute(ctx context.Context) error {
	if err := i.importSent(ctx); err != nil {
		return fmt.Errorf("import sent books: %w", err)
	}
	if err := i.importWishlist(ctx); err != nil {
		return fmt.Errorf("import wishlist: %w", err)
	}
	return nil
}

func (i *ImportSignalsUC) importSent(ctx context.Context) error {
	since := time.Time{}
	if value, err := i.signals.GetState(ctx, sentCursor); err != nil {
		return err
	} else if value != "" {
		if since, err = time.Parse(time.RFC3339, value); err != nil {
			return fmt.Errorf("parse %s: %w", sentCursor, err)
		}
	}

	for {
		books, err := i.sent.ListSince(ctx, since, pageSize)
		if err != nil {
			return err
		}
		for _, book := range books {
			title := models.CleanTitle(book.Title)
			if title == "" {
				title = models.CleanTitle(book.Filename)
			}
			if title == "" {
				continue
			}
			signal := &models.Signal{
				UserId: book.UserId,
				Ref:    "sent:" + book.SourceId,
				Source: models.SignalSent,
				Kind:   models.SignalBook,
				Title:  title,
			}
			if _, err := i.signals.AddSignal(ctx, signal); err != nil {
				return err
			}
		}
		if len(books) == 0 {
			return nil
		}
		last := books[len(books)-1].UpdatedAt
		if err := i.signals.SetState(ctx, sentCursor, last.UTC().Format(time.RFC3339)); err != nil {
			return err
		}
		// a page of books updated in the same second would come back forever
		if len(books) < pageSize || !last.After(since) {
			return nil
		}
		since = last
	}
}

func (i *ImportSignalsUC) importWishlist(ctx context.Context) error {
	items, err := i.wishlist.ListAll(ctx)
	if err != nil {
		return err
	}
	refs := make(map[string]bool, len(items))
	for _, item := range items {
		ref := "wish:" + strconv.FormatInt(item.Id, 10)
		refs[ref] = true
		signal := &models.Signal{
			UserId: item.UserId,
			Ref:    ref,
			Source: models.SignalWishlist,
			Kind:   item.Kind,
			Title:  item.Title,
			Author: item.Author,
		}
		// a suggestion picked from the catalog is already matched
		if key := item.WorkKey; item.Kind == models.SignalBook && key != "" {
			signal.MatchedKey = key
		} else if key := item.AuthorKey; item.Kind == models.SignalAuthor && key != "" {
			signal.MatchedKey = key
		}
		if _, err := i.signals.AddSignal(ctx, signal); err != nil {
			return err
		}
	}

	users, err := i.signals.RemoveWishlistExcept(ctx, refs)
	if err != nil {
		return err
	}
	return i.signals.MarkStale(ctx, users...)
}
