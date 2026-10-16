package models_test

import (
	"testing"
	"time"

	"github.com/ismelen/inkomi/back/transaction-manager/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSourceListFilter_InvalidInput_ReturnsError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		limit  int
		cursor string
		want   error
	}{
		{"zero limit", 0, "", models.ErrInvalidLimit},
		{"limit over max", 101, "", models.ErrInvalidLimit},
		{"not base64", 10, "%%%", models.ErrInvalidCursor},
		{"no separator", 10, "YWJj", models.ErrInvalidCursor},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Act
			filter, err := models.NewSourceListFilter(tt.limit, tt.cursor, nil)

			// Assert
			assert.ErrorIs(t, err, tt.want)
			assert.Nil(t, filter)
		})
	}
}

func TestNewSourceListFilter_EncodedCursor_DecodesIt(t *testing.T) {
	t.Parallel()

	// Arrange
	createdAt := time.Date(2026, 10, 16, 10, 30, 0, 0, time.UTC)
	cursor := models.NewSourceCursor(&models.Source{Id: "abc|def", CreatedAt: createdAt})

	// Act
	filter, err := models.NewSourceListFilter(100, cursor.Encode(), nil)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 100, filter.Limit)
	assert.Equal(t, cursor, filter.Cursor)
}
