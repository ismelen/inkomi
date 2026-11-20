package expandcatalog_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/expandcatalog"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/mocks"
)

func TestExpandCatalogUC_Execute_AuthorsAndSubjects_AddsCandidatesNoveltiesFirst(t *testing.T) {
	t.Parallel()

	// Arrange
	var calls []string
	source := &mocks.MockCatalogSource{
		NewWorksFn: func(ctx context.Context, subject, author string, since, limit int) ([]models.Work, error) {
			calls = append(calls, "new:"+subject+author)
			assert.Equal(t, 2025, since)
			return []models.Work{{Key: "N-" + subject + author}}, nil
		},
		WorksByAuthorFn: func(ctx context.Context, author string, limit int) ([]models.Work, error) {
			calls = append(calls, "author:"+author)
			return nil, errors.New("HTTP 500")
		},
		WorksBySubjectFn: func(ctx context.Context, subject string, limit int) ([]models.Work, error) {
			calls = append(calls, "subject:"+subject)
			return []models.Work{{Key: "S1"}, {Key: "S2"}}, nil
		},
	}
	var origins []models.Origin
	catalog := &mocks.MockCatalogRepository{
		TopAuthorKeysFn: func(ctx context.Context, limit int) ([]string, error) { return []string{"A1"}, nil },
		TopSubjectsFn:   func(ctx context.Context, limit int) ([]string, error) { return []string{"Thriller"}, nil },
		UpsertWorkFn: func(ctx context.Context, w *models.Work) (bool, error) {
			origins = append(origins, w.Origin)
			return w.Key != "S2", nil
		},
	}
	now := func() time.Time { return time.Date(2026, 11, 20, 0, 0, 0, 0, time.UTC) }
	uc := expandcatalog.NewExpandCatalogUC(source, catalog, now)

	// Act
	added, err := uc.Execute(context.Background(), models.NewBudget(10))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, []string{"new:A1", "new:Thriller", "author:A1", "subject:Thriller"}, calls)
	assert.Equal(t, 3, added, "S2 was already known")
	for _, origin := range origins {
		assert.Equal(t, models.OriginCandidate, origin)
	}
}

func TestExpandCatalogUC_Execute_SmallBudget_StopsEarly(t *testing.T) {
	t.Parallel()

	// Arrange
	calls := 0
	source := &mocks.MockCatalogSource{NewWorksFn: func(ctx context.Context, subject, author string, since, limit int) ([]models.Work, error) {
		calls++
		return nil, nil
	}}
	catalog := &mocks.MockCatalogRepository{
		TopAuthorKeysFn: func(ctx context.Context, limit int) ([]string, error) { return []string{"A1", "A2", "A3"}, nil },
	}
	uc := expandcatalog.NewExpandCatalogUC(source, catalog, time.Now)

	// Act
	_, err := uc.Execute(context.Background(), models.NewBudget(2))

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 2, calls)
}
