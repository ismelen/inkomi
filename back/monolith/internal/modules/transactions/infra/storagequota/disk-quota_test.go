package storagequota_test

import (
	"context"
	"database/sql"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/repositories"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/infra/storagequota"
	"github.com/ismelen/inkomi/back/monolith/migrations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

type usage map[string]int64

func (u usage) Usage(ids []string) int64 {
	var total int64
	for _, id := range ids {
		total += u[id]
	}
	return total
}

func newDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	_, err = db.Exec(migrations.Transactions)
	require.NoError(t, err)
	t.Cleanup(func() {
		// in-memory database
		_ = db.Close()
	})
	return db
}

func newQuota(t *testing.T, free uint64, used usage, now func() time.Time) (*storagequota.DiskQuota, *sql.DB) {
	t.Helper()
	db := newDB(t)
	repo := repositories.NewSQLiteStorageReservationRepository(db)
	return storagequota.NewDiskQuota(repo, used, func() (uint64, bool) { return free, true }, 100, now), db
}

func reservation(root string, bytes int64) []models.StorageReservation {
	return []models.StorageReservation{{RootId: root, UserId: 1, Bytes: bytes}}
}

func TestDiskQuota_Reserve_Fits_Records(t *testing.T) {
	t.Parallel()

	// Arrange: 1000 free, 100 of reserve
	q, db := newQuota(t, 1000, usage{}, time.Now)

	// Act
	err := q.Reserve(context.Background(), reservation("a", 900))

	// Assert
	require.NoError(t, err)
	var n int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM storage_reservation`).Scan(&n))
	assert.Equal(t, 1, n)
}

func TestDiskQuota_Reserve_DoesNotFit_ReturnsInsufficientStorage(t *testing.T) {
	t.Parallel()

	// Arrange
	q, _ := newQuota(t, 1000, usage{}, time.Now)

	// Act
	err := q.Reserve(context.Background(), reservation("a", 901))

	// Assert
	assert.ErrorIs(t, err, models.ErrInsufficientStorage)
}

func TestDiskQuota_Reserve_CountsWhatActiveUploadsStillNeed(t *testing.T) {
	t.Parallel()

	// Arrange: an upload reserved 600 and its files already take 400 of them, which statfs
	// already counts as used, so 200 are still pending
	used := usage{}
	q, db := newQuota(t, 1000, used, time.Now)
	ctx := context.Background()
	size := int64(10)
	id, err := repositories.NewSQLiteSourceRepository(db).Create(ctx, &models.Source{UserId: 1, Title: "a", Filename: "a.cbz", Size: &size, IsFile: true})
	require.NoError(t, err)
	used[id] = 400
	require.NoError(t, q.Reserve(ctx, reservation(id, 600)))

	// Act: free 1000 - pending 200 - reserve 100 = 700
	tooBig := q.Reserve(ctx, reservation("b", 701))
	fits := q.Reserve(ctx, reservation("c", 700))

	// Assert
	assert.ErrorIs(t, tooBig, models.ErrInsufficientStorage)
	assert.NoError(t, fits)
}

func TestDiskQuota_Exclusive_ConcurrentUploadsDontOvercommit(t *testing.T) {
	t.Parallel()

	// Arrange: room for one 500-byte upload, not two
	q, _ := newQuota(t, 1000, usage{}, time.Now)
	var accepted atomic.Int32
	var wg sync.WaitGroup

	// Act
	for _, root := range []string{"a", "b"} {
		wg.Go(func() {
			err := q.Exclusive(func() error { return q.Reserve(context.Background(), reservation(root, 500)) })
			if err == nil {
				accepted.Add(1)
			}
		})
	}
	wg.Wait()

	// Assert
	assert.Equal(t, int32(1), accepted.Load())
}

func TestDiskQuota_ReleaseFinished_FreesTheTreeOnceDone(t *testing.T) {
	t.Parallel()

	// Arrange: a loose file that finishes
	q, db := newQuota(t, 1000, usage{}, time.Now)
	ctx := context.Background()
	sources := repositories.NewSQLiteSourceRepository(db)
	size := int64(10)
	id, err := sources.Create(ctx, &models.Source{UserId: 1, Title: "a", Filename: "a.cbz", Size: &size, IsFile: true})
	require.NoError(t, err)
	require.NoError(t, q.Reserve(ctx, reservation(id, 800)))

	// Act & Assert: still processing, nothing else fits
	require.NoError(t, sources.UpdateStatus(ctx, id, 1, models.SourceStatusProcessing))
	require.NoError(t, q.ReleaseFinished(ctx, id, 1))
	assert.ErrorIs(t, q.Reserve(ctx, reservation("b", 800)), models.ErrInsufficientStorage)

	require.NoError(t, sources.UpdateStatus(ctx, id, 1, models.SourceStatusSent))
	require.NoError(t, q.ReleaseFinished(ctx, id, 1))
	assert.NoError(t, q.Reserve(ctx, reservation("b", 800)))
}

func TestDiskQuota_Sweep_ReleasesUnconfirmedUploads(t *testing.T) {
	t.Parallel()

	// Arrange: an upload never confirmed, seen an hour later
	later := time.Now().Add(time.Hour)
	q, db := newQuota(t, 1000, usage{}, func() time.Time { return later })
	ctx := context.Background()
	sources := repositories.NewSQLiteSourceRepository(db)
	size := int64(10)
	id, err := sources.Create(ctx, &models.Source{UserId: 1, Title: "a", Filename: "a.cbz", Size: &size, IsFile: true})
	require.NoError(t, err)
	require.NoError(t, q.Reserve(ctx, reservation(id, 800)))

	// Act
	require.NoError(t, q.Sweep(ctx))

	// Assert
	assert.NoError(t, q.Reserve(ctx, reservation("b", 800)))
}
