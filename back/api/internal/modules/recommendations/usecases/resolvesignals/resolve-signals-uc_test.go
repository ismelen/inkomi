package resolvesignals_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/mocks"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/resolvesignals"
)

type resolution struct {
	id     int64
	status models.SignalStatus
	key    string
}

func TestResolveSignalsUC_Execute_Signals_MatchesThemAgainstTheCatalog(t *testing.T) {
	t.Parallel()

	// Arrange
	pending := []models.Signal{
		{Id: 1, UserId: 7, Kind: models.SignalBook, Title: "La nena"},
		{Id: 2, UserId: 7, Kind: models.SignalBook, Title: "Dune", MatchedKey: "OL2W"},
		{Id: 3, UserId: 8, Kind: models.SignalAuthor, Title: "Carmen Mola"},
		{Id: 4, UserId: 8, Kind: models.SignalBook, Title: "Nothing like it", Attempts: 2},
		{Id: 5, UserId: 8, Kind: models.SignalBook, Title: "Broken"},
	}
	source := &mocks.MockCatalogSource{
		MatchBookFn: func(ctx context.Context, title, author string) (*models.Work, error) {
			switch title {
			case "La nena":
				return &models.Work{Key: "OL1W", Title: "La nena"}, nil
			case "Broken":
				return nil, errors.New("HTTP 503")
			}
			return nil, nil
		},
		WorkFn: func(ctx context.Context, key string) (*models.Work, error) {
			return &models.Work{Key: key, Title: "Dune"}, nil
		},
		MatchAuthorFn: func(ctx context.Context, name string) (string, string, error) { return "OL9A", name, nil },
		DescriptionFn: func(ctx context.Context, key string) (string, error) { return "about " + key, nil },
	}
	var upserted []models.Work
	catalog := &mocks.MockCatalogRepository{UpsertWorkFn: func(ctx context.Context, w *models.Work) (bool, error) {
		upserted = append(upserted, *w)
		return true, nil
	}}
	var resolved []resolution
	var stale []int
	signals := &mocks.MockSignalRepository{
		ListPendingFn: func(ctx context.Context, limit, maxAttempts int) ([]models.Signal, error) { return pending, nil },
		ResolveFn: func(ctx context.Context, id int64, status models.SignalStatus, key string) error {
			resolved = append(resolved, resolution{id, status, key})
			return nil
		},
		MarkStaleFn: func(ctx context.Context, ids ...int) error {
			stale = append(stale, ids...)
			return nil
		},
	}
	uc := resolvesignals.NewResolveSignalsUC(source, catalog, signals)

	// Act
	matched, err := uc.Execute(context.Background(), models.NewBudget(100))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 3, matched)
	assert.Equal(t, []resolution{
		{1, models.SignalMatched, "OL1W"},
		{2, models.SignalMatched, "OL2W"},
		{3, models.SignalMatched, "OL9A"},
		{4, models.SignalUnmatched, ""},
		{5, models.SignalPending, ""},
	}, resolved)
	require.Len(t, upserted, 2)
	assert.Equal(t, models.OriginUser, upserted[0].Origin)
	assert.Equal(t, "about OL1W", upserted[0].Description)
	assert.Equal(t, []int{7, 7, 8}, stale)
}

func TestResolveSignalsUC_Execute_NoBudget_LeavesTheRestPending(t *testing.T) {
	t.Parallel()

	// Arrange
	calls := 0
	source := &mocks.MockCatalogSource{MatchBookFn: func(ctx context.Context, title, author string) (*models.Work, error) {
		calls++
		return &models.Work{Key: "OL1W"}, nil
	}}
	signals := &mocks.MockSignalRepository{ListPendingFn: func(ctx context.Context, limit, maxAttempts int) ([]models.Signal, error) {
		return []models.Signal{{Id: 1, Kind: models.SignalBook}, {Id: 2, Kind: models.SignalBook}}, nil
	}}
	uc := resolvesignals.NewResolveSignalsUC(source, &mocks.MockCatalogRepository{}, signals)

	// Act
	matched, err := uc.Execute(context.Background(), models.NewBudget(3))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, matched, "a book takes two requests")
	assert.Equal(t, 1, calls)
}
