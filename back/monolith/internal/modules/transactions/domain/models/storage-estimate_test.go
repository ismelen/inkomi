package models_test

import (
	"math"
	"testing"

	"github.com/ismelen/inkomi/back/monolith/internal/modules/transactions/domain/models"
	"github.com/stretchr/testify/assert"
)

const mb = int64(1 << 20)

func TestEstimateStoragePeak_Phases(t *testing.T) {
	t.Parallel()

	opts := models.StorageEstimateOptions{EpubRatio: 1.2, LibraryEstimate: 100 * mb, Concurrency: 2, MaxJoinBlock: 200 * mb}
	tenChapters := make([]models.StorageItem, 10)
	for i := range tenChapters {
		tenChapters[i] = models.StorageItem{Size: 10 * mb}
	}

	cases := []struct {
		name string
		req  models.StorageRequest
		want int64
	}{
		{"Empty", models.StorageRequest{}, 0},
		// 12 MB epub, plus the epub being written while its source is still there
		{"One manga file", models.StorageRequest{Items: []models.StorageItem{{Size: 10 * mb}}}, 24 * mb},
		// already an epub: no conversion output besides itself
		{"One epub", models.StorageRequest{Items: []models.StorageItem{{Size: 10 * mb, IsEpub: true}}}, 20 * mb},
		{"Library item", models.StorageRequest{Items: []models.StorageItem{{IsLibrary: true}}}, 200 * mb},
		// conversion: 10×12 + 2 in flight × 12 = 144 MB
		{"Folder without join", models.StorageRequest{Items: tenChapters}, 144 * mb},
		// join: 120 MB of epubs + a 120 MB block (smaller than the 200 MB limit)
		{"Folder joined", models.StorageRequest{Items: tenChapters, Join: true}, 240 * mb},
		// kepub of the joined block: 120 + 120, same as the join
		{"Folder joined to kepub", models.StorageRequest{Items: tenChapters, Join: true, Kepub: true}, 240 * mb},
		// kepub of loose epubs: 120 + 12, below the conversion peak
		{"Folder to kepub", models.StorageRequest{Items: tenChapters, Kepub: true}, 144 * mb},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Act
			got := models.EstimateStoragePeak(tc.req, opts)

			// Assert
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEstimateStoragePeak_BigJoin_UsesBlockLimit(t *testing.T) {
	t.Parallel()

	// Arrange: 84 chapters of 11 MB, like a real folder
	items := make([]models.StorageItem, 84)
	for i := range items {
		items[i] = models.StorageItem{Size: 11 * mb}
	}
	opts := models.StorageEstimateOptions{EpubRatio: 1.2, Concurrency: 2, MaxJoinBlock: 200 * mb}

	// Act
	got := models.EstimateStoragePeak(models.StorageRequest{Items: items, Join: true, Kepub: true}, opts)

	// Assert: every epub (84 × 13.2 MB) plus one 200 MB block, not twice the folder
	epub := int64(math.Ceil(float64(11*mb) * 1.2))
	assert.Equal(t, 84*epub+200*mb, got)
}
