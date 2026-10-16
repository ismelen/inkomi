package listuploads_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/listuploads"
	"github.com/ismelen/inkomi/back/transaction-manager/internal/usecases/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var now = time.Date(2026, 10, 16, 10, 0, 0, 0, time.UTC)

func trees(n int) []models.SourceTree {
	result := make([]models.SourceTree, n)
	for i := range result {
		result[i] = models.SourceTree{Source: models.Source{
			Id:        fmt.Sprintf("source-%d", i),
			CreatedAt: now.Add(-time.Duration(i) * time.Minute),
		}}
	}
	return result
}

func TestListUploadsUC_Execute_MoreThanLimit_ReturnsPageWithCursor(t *testing.T) {
	t.Parallel()

	// Arrange
	var gotFilter models.SourceListFilter
	repo := &mocks.MockSourceRepository{
		NowFn: func(ctx context.Context) (time.Time, error) { return now, nil },
		ListTreesByUserIdFn: func(ctx context.Context, userId int, filter models.SourceListFilter) ([]models.SourceTree, error) {
			gotFilter = filter
			return trees(3), nil
		},
	}
	uc := listuploads.NewListUploadsUC(repo)

	// Act
	page, err := uc.Execute(context.Background(), 1, models.SourceListFilter{Limit: 2})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 3, gotFilter.Limit)
	assert.Len(t, page.Trees, 2)
	require.NotNil(t, page.NextCursor)
	assert.Equal(t, "source-1", page.NextCursor.Id)
	assert.Equal(t, now.Add(-time.Minute), page.NextCursor.CreatedAt)
	assert.Equal(t, now, page.SyncedAt)
}

func TestListUploadsUC_Execute_LastPage_ReturnsNilCursor(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := &mocks.MockSourceRepository{
		ListTreesByUserIdFn: func(ctx context.Context, userId int, filter models.SourceListFilter) ([]models.SourceTree, error) {
			return trees(2), nil
		},
	}
	uc := listuploads.NewListUploadsUC(repo)

	// Act
	page, err := uc.Execute(context.Background(), 1, models.SourceListFilter{Limit: 2})

	// Assert
	require.NoError(t, err)
	assert.Len(t, page.Trees, 2)
	assert.Nil(t, page.NextCursor)
}

func TestListUploadsUC_Execute_ForwardsUserAndFilter(t *testing.T) {
	t.Parallel()

	// Arrange
	since := now.Add(-time.Hour)
	cursor := &models.SourceCursor{CreatedAt: now, Id: "x"}
	var gotUserId int
	var gotFilter models.SourceListFilter
	repo := &mocks.MockSourceRepository{
		ListTreesByUserIdFn: func(ctx context.Context, userId int, filter models.SourceListFilter) ([]models.SourceTree, error) {
			gotUserId, gotFilter = userId, filter
			return nil, nil
		},
	}
	uc := listuploads.NewListUploadsUC(repo)

	// Act
	_, err := uc.Execute(context.Background(), 7, models.SourceListFilter{Limit: 10, Cursor: cursor, UpdatedSince: &since})

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 7, gotUserId)
	assert.Equal(t, cursor, gotFilter.Cursor)
	assert.Equal(t, &since, gotFilter.UpdatedSince)
}

func TestListUploadsUC_Execute_RepositoryFails_ReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		repo *mocks.MockSourceRepository
	}{
		{"now", &mocks.MockSourceRepository{
			NowFn: func(ctx context.Context) (time.Time, error) { return time.Time{}, errors.New("boom") },
		}},
		{"list", &mocks.MockSourceRepository{
			ListTreesByUserIdFn: func(ctx context.Context, userId int, filter models.SourceListFilter) ([]models.SourceTree, error) {
				return nil, errors.New("boom")
			},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			uc := listuploads.NewListUploadsUC(tt.repo)

			// Act
			page, err := uc.Execute(context.Background(), 1, models.SourceListFilter{Limit: 1})

			// Assert
			assert.Error(t, err)
			assert.Nil(t, page)
		})
	}
}
