package importsignals_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/importsignals"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/mocks"
)

func TestImportSignalsUC_Execute_SentAndWished_AddsSignalsAndMovesTheCursor(t *testing.T) {
	t.Parallel()

	// Arrange
	sentAt := time.Date(2026, 11, 20, 10, 0, 0, 0, time.UTC)
	var gotSince time.Time
	sent := &mocks.MockSentBooks{ListSinceFn: func(ctx context.Context, since time.Time, limit int) ([]models.SentBook, error) {
		gotSince = since
		return []models.SentBook{
			{SourceId: "s1", UserId: 1, Title: "La nena.epub", UpdatedAt: sentAt},
			{SourceId: "s2", UserId: 1, Title: "", Filename: "", UpdatedAt: sentAt},
		}, nil
	}}
	wishlist := &mocks.MockWishlist{ListAllFn: func(ctx context.Context) ([]models.WishedItem, error) {
		return []models.WishedItem{
			{Id: 4, UserId: 2, Kind: models.SignalBook, Title: "Dune", WorkKey: "OL1W"},
			{Id: 5, UserId: 2, Kind: models.SignalAuthor, Title: "Carmen Mola"},
		}, nil
	}}
	var added []models.Signal
	state := map[string]string{"sent_cursor": "2026-11-01T00:00:00Z"}
	var keptRefs map[string]bool
	var stale []int
	signals := &mocks.MockSignalRepository{
		AddSignalFn: func(ctx context.Context, s *models.Signal) (bool, error) {
			added = append(added, *s)
			return true, nil
		},
		GetStateFn: func(ctx context.Context, name string) (string, error) { return state[name], nil },
		SetStateFn: func(ctx context.Context, name, value string) error {
			state[name] = value
			return nil
		},
		RemoveWishlistExceptFn: func(ctx context.Context, refs map[string]bool) ([]int, error) {
			keptRefs = refs
			return []int{9}, nil
		},
		MarkStaleFn: func(ctx context.Context, ids ...int) error {
			stale = append(stale, ids...)
			return nil
		},
	}
	uc := importsignals.NewImportSignalsUC(sent, wishlist, signals)

	// Act
	err := uc.Execute(context.Background())

	// Assert
	require.NoError(t, err)
	assert.Equal(t, time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC), gotSince)
	require.Len(t, added, 3, "the book without any title is skipped")
	assert.Equal(t, models.Signal{UserId: 1, Ref: "sent:s1", Source: models.SignalSent, Kind: models.SignalBook, Title: "La nena"}, added[0])
	assert.Equal(t, "OL1W", added[1].MatchedKey, "a picked suggestion is already matched")
	assert.Equal(t, models.SignalWishlist, added[1].Source)
	assert.Empty(t, added[2].MatchedKey)
	assert.Equal(t, "2026-11-20T10:00:00Z", state["sent_cursor"])
	assert.Equal(t, map[string]bool{"wish:4": true, "wish:5": true}, keptRefs)
	assert.Equal(t, []int{9}, stale)
}
