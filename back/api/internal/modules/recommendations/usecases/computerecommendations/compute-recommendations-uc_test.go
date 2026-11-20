package computerecommendations_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/computerecommendations"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/mocks"
)

func vector(key string, year int, authors []string, x, y float32) models.WorkVector {
	return models.WorkVector{Key: key, AuthorKeys: authors, FirstPublishYear: year, Embedding: models.Embedding{x, y}}
}

func withTitle(v models.WorkVector, title string) models.WorkVector {
	v.Title = title
	return v
}

func withLanguages(v models.WorkVector, languages ...string) models.WorkVector {
	v.Languages = languages
	return v
}

func keys(recs []models.Recommendation, section models.Section) []string {
	var result []string
	for _, rec := range recs {
		if rec.Section == section {
			result = append(result, rec.Work.Key)
		}
	}
	return result
}

func TestComputeRecommendationsUC_Execute_Taste_RanksEverySection(t *testing.T) {
	t.Parallel()

	// Arrange: the user sent OWN (by MOLA) and wishes the author Z
	vectors := []models.WorkVector{
		withTitle(vector("OWN", 2010, []string{"MOLA"}, 1, 0), "La nena"),
		withTitle(vector("EDITION", 2025, []string{"PUB"}, 1, 0), "La Nena (Edición limitada)"),
		withLanguages(vector("SPA", 1990, []string{"SPA"}, 0.2, 0.8), "spa"),
		vector("ENG", 1990, []string{"ENG"}, 0.2, 0.8),
		vector("T1", 2000, []string{"MOLA"}, 0.99, 0.1),
		vector("T2", 2026, []string{"MOLA"}, 0.98, 0.15),
		vector("T3", 2005, []string{"MOLA"}, 0.97, 0.2),
		vector("T4", 2001, []string{"OTHER"}, 0.9, 0.3),
		vector("Z1", 2026, []string{"Z"}, 0.5, 0.5),
		vector("FAR", 2026, []string{"FAR"}, -1, 0),
		vector("GONE", 2000, []string{"X"}, 1, 0),
	}
	catalog := &mocks.MockCatalogRepository{ListVectorsFn: func(ctx context.Context) ([]models.WorkVector, error) { return vectors, nil }}
	signals := &mocks.MockSignalRepository{
		ListMatchedFn: func(ctx context.Context, userId int) ([]models.Signal, error) {
			return []models.Signal{
				{Kind: models.SignalBook, Source: models.SignalSent, MatchedKey: "OWN"},
				{Kind: models.SignalAuthor, Source: models.SignalWishlist, MatchedKey: "Z"},
			}, nil
		},
		ListDismissedFn: func(ctx context.Context, userId int) (map[string]bool, error) {
			return map[string]bool{"GONE": true}, nil
		},
	}
	var stored []models.Recommendation
	recs := &mocks.MockRecommendationRepository{ReplaceFn: func(ctx context.Context, userId int, r []models.Recommendation) error {
		stored = r
		return nil
	}}
	now := func() time.Time { return time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC) }
	uc := computerecommendations.NewComputeRecommendationsUC(catalog, signals, recs, []string{"spa"}, now)

	// Act
	err := uc.Execute(context.Background(), []int{1})

	// Assert: the profile is OWN pulled a bit towards Z1, so T3 > T4 > T2 > T1 > Z1 > FAR
	require.NoError(t, err)
	assert.Equal(t, []string{"T3", "T4", "T2", "Z1", "SPA", "ENG", "FAR"}, keys(stored, models.SectionForYou),
		"no own, dismissed or other editions of own works, at most two per author (T1 is the third of MOLA) and the preferred language first on a tie")
	assert.Equal(t, []string{"T2", "Z1", "FAR"}, keys(stored, models.SectionNew), "only recent works")
	assert.Equal(t, []string{"T4", "SPA", "ENG", "FAR"}, keys(stored, models.SectionDiscover), "no known or wished authors")
	for _, rec := range stored {
		require.NotNil(t, rec.Reason)
		assert.Equal(t, "OWN", rec.Reason.Key)
	}
}

func TestComputeRecommendationsUC_Execute_NoSignals_ClearsTheRecommendations(t *testing.T) {
	t.Parallel()

	// Arrange
	replaced := false
	var got []models.Recommendation
	uc := computerecommendations.NewComputeRecommendationsUC(
		&mocks.MockCatalogRepository{ListVectorsFn: func(ctx context.Context) ([]models.WorkVector, error) {
			return []models.WorkVector{vector("A", 2000, nil, 1, 0)}, nil
		}},
		&mocks.MockSignalRepository{},
		&mocks.MockRecommendationRepository{ReplaceFn: func(ctx context.Context, userId int, r []models.Recommendation) error {
			replaced, got = true, r
			return nil
		}},
		nil,
		time.Now,
	)

	// Act
	err := uc.Execute(context.Background(), []int{1})

	// Assert
	require.NoError(t, err)
	assert.True(t, replaced)
	assert.Empty(t, got)
}
