package refreshcatalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/computerecommendations"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/embedworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/expandcatalog"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/importsignals"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/refreshcatalog"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/resolvesignals"
)

type spy struct {
	expanded bool
	pruned   bool
	ranked   []int
}

func newRefresh(s *spy, embedded int) *refreshcatalog.RefreshCatalogUC {
	unembedded := embedded
	source := &mocks.MockCatalogSource{}
	catalog := &mocks.MockCatalogRepository{
		TopSubjectsFn: func(ctx context.Context, limit int) ([]string, error) {
			s.expanded = true
			return nil, nil
		},
		PruneCandidatesFn: func(ctx context.Context, before time.Time) (int, error) {
			s.pruned = true
			return 0, nil
		},
		ListUnembeddedFn: func(ctx context.Context, limit int) ([]models.Work, error) {
			if unembedded == 0 {
				return nil, nil
			}
			unembedded = 0
			return make([]models.Work, embedded), nil
		},
		ListVectorsFn: func(ctx context.Context) ([]models.WorkVector, error) { return nil, nil },
	}
	signals := &mocks.MockSignalRepository{
		TakeStaleFn: func(ctx context.Context) ([]int, error) { return []int{1}, nil },
		ListUsersFn: func(ctx context.Context) ([]int, error) { return []int{1, 2, 3}, nil },
	}
	recs := &mocks.MockRecommendationRepository{ReplaceFn: func(ctx context.Context, userId int, r []models.Recommendation) error {
		s.ranked = append(s.ranked, userId)
		return nil
	}}
	return refreshcatalog.NewRefreshCatalogUC(
		importsignals.NewImportSignalsUC(&mocks.MockSentBooks{}, &mocks.MockWishlist{}, signals),
		resolvesignals.NewResolveSignalsUC(source, catalog, signals),
		expandcatalog.NewExpandCatalogUC(source, catalog, time.Now),
		embedworks.NewEmbedWorksUC(&mocks.MockEmbedder{}, catalog, nil),
		computerecommendations.NewComputeRecommendationsUC(catalog, signals, recs, nil, time.Now),
		catalog, signals, 100, time.Now,
	)
}

func TestRefreshCatalogUC_Execute_Passes_FullGrowsTheCatalogAndRanksEveryone(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		full     bool
		embedded int
		expanded bool
		ranked   []int
	}{
		{"Quick", false, 2, false, []int{1}},
		{"Full without changes", true, 0, true, []int{1}},
		{"Full with new embeddings", true, 2, true, []int{1, 2, 3}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			s := &spy{}
			uc := newRefresh(s, tc.embedded)

			// Act
			err := uc.Execute(context.Background(), tc.full)

			// Assert
			require.NoError(t, err)
			assert.Equal(t, tc.expanded, s.expanded)
			assert.Equal(t, tc.expanded, s.pruned)
			assert.Equal(t, tc.ranked, s.ranked)
		})
	}
}
