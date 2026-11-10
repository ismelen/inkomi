package canceltransaction

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
)

type CancelTransactionUC struct {
	txManager    ports.ITxManager
	sourceRepo   ports.ISourceRepository
	reservations ports.IStorageReservationRepository
	registry     ports.ICancelRegistry
	files        ports.IFileCleaner
	joins        ports.IJoinGroups
	hub          ports.ISocketHub
}

func NewCancelTransactionUC(
	txManager ports.ITxManager,
	sourceRepo ports.ISourceRepository,
	reservations ports.IStorageReservationRepository,
	registry ports.ICancelRegistry,
	files ports.IFileCleaner,
	joins ports.IJoinGroups,
	hub ports.ISocketHub,
) *CancelTransactionUC {
	return &CancelTransactionUC{
		txManager:    txManager,
		sourceRepo:   sourceRepo,
		reservations: reservations,
		registry:     registry,
		files:        files,
		joins:        joins,
		hub:          hub,
	}
}

// Execute cancels the upload rooted at rootId: what wasn't finished becomes cancelled, its jobs
// stop, its files are deleted and its disk reservation is freed. What was already sent stays.
func (c *CancelTransactionUC) Execute(ctx context.Context, userId int, rootId string) error {
	root, err := c.sourceRepo.GetByIdAndUserId(ctx, rootId, userId)
	if err != nil {
		return err
	}
	if root.ParentId != nil {
		return models.ErrSourceNotCancellable
	}

	var cancelled []string
	err = c.txManager.ExecuteTx(ctx, func(txCtx context.Context) error {
		ids, err := c.sourceRepo.CancelTree(txCtx, rootId, userId)
		if err != nil {
			return err
		}
		cancelled = ids
		return c.reservations.Release(txCtx, rootId)
	})
	if err != nil {
		return err
	}
	if len(cancelled) == 0 {
		return models.ErrSourceNotCancellable
	}

	treeIds, err := c.reservations.TreeIds(ctx, rootId)
	if err != nil {
		return err
	}
	if err := c.registry.Cancel(ctx, treeIds); err != nil {
		return err
	}
	log.Printf("upload cancel user %d: %s (%d sources)", userId, rootId, len(cancelled))

	// the upload is already cancelled: what's left is cleanup, so it's only logged
	var errs []error
	for _, id := range cancelled {
		if err := c.files.DeleteAll(id); err != nil {
			errs = append(errs, fmt.Errorf("delete files of %s: %w", id, err))
		}
		if err := c.hub.Send(userId, map[string]any{"id": id, "status": models.SourceStatusCancelled}); err != nil &&
			!errors.Is(err, models.ErrClientNotConnected) {
			errs = append(errs, fmt.Errorf("notify %s: %w", id, err))
		}
	}
	if err := c.joins.Clear(ctx, rootId); err != nil {
		errs = append(errs, fmt.Errorf("clear join group: %w", err))
	}
	if err := errors.Join(errs...); err != nil {
		log.Printf("upload cancel %s: %v", rootId, err)
	}
	return nil
}
