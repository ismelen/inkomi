package listsentbooks_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/usecases/listsentbooks"
	"github.com/ismelen/inkomi/back/api/internal/modules/transactions/usecases/mocks"
)

func TestListSentBooksUC_Execute_Limit_ClampsIt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		limit int
		want  int
	}{
		{"Within range", 50, 50},
		{"Zero", 0, 500},
		{"Too big", 10000, 500},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Arrange
			since := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
			var gotSince time.Time
			var gotLimit int
			uc := listsentbooks.NewListSentBooksUC(&mocks.MockSourceRepository{
				ListSentBooksFn: func(ctx context.Context, s time.Time, limit int) ([]models.SentBook, error) {
					gotSince, gotLimit = s, limit
					return []models.SentBook{{SourceId: "a"}}, nil
				},
			})

			// Act
			books, err := uc.Execute(context.Background(), since, tc.limit)

			// Assert
			require.NoError(t, err)
			assert.Len(t, books, 1)
			assert.Equal(t, since, gotSince)
			assert.Equal(t, tc.want, gotLimit)
		})
	}
}
