package canceltransaction_test

import (
	"context"
	"testing"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/ports"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/canceltransaction"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type reservations struct {
	ports.IStorageReservationRepository
	released []string
	tree     []string
}

func (r *reservations) Release(ctx context.Context, rootId string) error {
	r.released = append(r.released, rootId)
	return nil
}

func (r *reservations) TreeIds(ctx context.Context, rootId string) ([]string, error) {
	return r.tree, nil
}

type recorder struct {
	cancelled []string
	deleted   []string
	cleared   []string
}

func (r *recorder) Cancel(ctx context.Context, ids []string) error {
	r.cancelled = append(r.cancelled, ids...)
	return nil
}

func (r *recorder) DeleteAll(id string) error {
	r.deleted = append(r.deleted, id)
	return nil
}

func (r *recorder) Clear(ctx context.Context, parentId string) error {
	r.cleared = append(r.cleared, parentId)
	return nil
}

type fixture struct {
	repo  *mocks.MockSourceRepository
	res   *reservations
	rec   *recorder
	sent  []any
	uc    *canceltransaction.CancelTransactionUC
	unfin []string
}

func newFixture(root *models.Source, unfinished []string) *fixture {
	f := &fixture{res: &reservations{tree: []string{"root", "ch-1", "ch-2"}}, rec: &recorder{}, unfin: unfinished}
	f.repo = &mocks.MockSourceRepository{
		GetByIdAndUserIdFn: func(ctx context.Context, id string, userId int) (*models.Source, error) {
			if root == nil || userId != root.UserId {
				return nil, models.ErrSourceNotFound
			}
			return root, nil
		},
		CancelTreeFn: func(ctx context.Context, rootId string, userId int) ([]string, error) {
			return f.unfin, nil
		},
	}
	hub := &mocks.MockSocketHub{SendFn: func(id int, v any) error {
		f.sent = append(f.sent, v)
		return nil
	}}
	f.uc = canceltransaction.NewCancelTransactionUC(&mocks.MockTxManager{}, f.repo, f.res, f.rec, f.rec, f.rec, hub)
	return f
}

func TestCancelTransactionUC_Execute_PendingTree_CancelsEverything(t *testing.T) {
	t.Parallel()

	// Arrange: ch-1 was already sent, the root and ch-2 weren't finished
	f := newFixture(&models.Source{Id: "root", UserId: 7}, []string{"root", "ch-2"})

	// Act
	err := f.uc.Execute(context.Background(), 7, "root")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"root"}, f.res.released)
	assert.Equal(t, []string{"root", "ch-1", "ch-2"}, f.rec.cancelled, "every job of the tree stops")
	assert.Equal(t, []string{"root", "ch-2"}, f.rec.deleted, "the sent chapter keeps its file")
	assert.Equal(t, []string{"root"}, f.rec.cleared)
	assert.Len(t, f.sent, 2)
}

func TestCancelTransactionUC_Execute_FinishedTree_ReturnsNotCancellable(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(&models.Source{Id: "root", UserId: 7}, nil)

	// Act
	err := f.uc.Execute(context.Background(), 7, "root")

	// Assert
	assert.ErrorIs(t, err, models.ErrSourceNotCancellable)
	assert.Empty(t, f.rec.cancelled)
}

func TestCancelTransactionUC_Execute_NotARoot_ReturnsNotCancellable(t *testing.T) {
	t.Parallel()

	// Arrange
	parent := "root"
	f := newFixture(&models.Source{Id: "ch-1", UserId: 7, ParentId: &parent}, []string{"ch-1"})

	// Act
	err := f.uc.Execute(context.Background(), 7, "ch-1")

	// Assert
	assert.ErrorIs(t, err, models.ErrSourceNotCancellable)
}

func TestCancelTransactionUC_Execute_OtherUser_ReturnsNotFound(t *testing.T) {
	t.Parallel()

	// Arrange
	f := newFixture(&models.Source{Id: "root", UserId: 7}, []string{"root"})

	// Act
	err := f.uc.Execute(context.Background(), 8, "root")

	// Assert
	assert.ErrorIs(t, err, models.ErrSourceNotFound)
}
