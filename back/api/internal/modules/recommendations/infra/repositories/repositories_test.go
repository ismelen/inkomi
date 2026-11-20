package repositories_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/infra/repositories"
	"github.com/ismelen/inkomi/back/api/migrations"

	_ "modernc.org/sqlite"
)

func setupDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.Exec(migrations.Discover)
	require.NoError(t, err)
	return db
}

func matchedSignal(t *testing.T, signals *repositories.SQLiteSignalRepository, userId int, ref string, kind models.SignalKind, key string) {
	t.Helper()

	ctx := context.Background()
	s := &models.Signal{UserId: userId, Ref: ref, Source: models.SignalSent, Kind: kind, Title: ref}
	_, err := signals.AddSignal(ctx, s)
	require.NoError(t, err)
	pending, err := signals.ListPending(ctx, 100, 3)
	require.NoError(t, err)
	for _, p := range pending {
		if p.Ref == ref && p.UserId == userId {
			require.NoError(t, signals.Resolve(ctx, p.Id, models.SignalMatched, key))
		}
	}
}

func TestSQLiteCatalogRepository_UpsertWork_RicherData_MergesAndDropsStaleEmbeddings(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := repositories.NewSQLiteCatalogRepository(setupDB(t))
	user := &models.Work{Key: "OL1W", Title: "La nena", Authors: []string{"Carmen Mola"}, AuthorKeys: []string{"OL9A"},
		Subjects: []string{"Thriller", "Madrid"}, FirstPublishYear: 2020, CoverId: 5, Description: "Una novela", Origin: models.OriginUser}

	// Act
	created, err := repo.UpsertWork(ctx, user)
	require.NoError(t, err)
	require.NoError(t, repo.SaveEmbedding(ctx, "OL1W", models.Embedding{0.5, -0.25}))
	again, err := repo.UpsertWork(ctx, &models.Work{Key: "OL1W", Title: "La nena", Subjects: []string{"Thriller"}, Origin: models.OriginCandidate})
	require.NoError(t, err)
	vectorsAfterPoorer, err := repo.ListVectors(ctx)
	require.NoError(t, err)
	_, err = repo.UpsertWork(ctx, &models.Work{Key: "OL1W", Title: "La nena", Subjects: []string{"Thriller", "Madrid", "Police"}, Origin: models.OriginCandidate})
	require.NoError(t, err)
	vectorsAfterRicher, err := repo.ListVectors(ctx)
	require.NoError(t, err)
	works, err := repo.GetWorks(ctx, []string{"OL1W", "missing"})
	require.NoError(t, err)

	// Assert
	assert.True(t, created)
	assert.False(t, again)
	require.Len(t, vectorsAfterPoorer, 1, "poorer data keeps the embedding")
	assert.Equal(t, models.Embedding{0.5, -0.25}, vectorsAfterPoorer[0].Embedding)
	assert.Equal(t, []string{"OL9A"}, vectorsAfterPoorer[0].AuthorKeys)
	assert.Empty(t, vectorsAfterRicher, "new subjects change the passage")
	require.Contains(t, works, "OL1W")
	work := works["OL1W"]
	assert.Equal(t, models.OriginUser, work.Origin, "a user work stays a user work")
	assert.Equal(t, []string{"Carmen Mola"}, work.Authors)
	assert.Equal(t, []string{"Thriller", "Madrid", "Police"}, work.Subjects)
	assert.Equal(t, "Una novela", work.Description)
	assert.Equal(t, 2020, work.FirstPublishYear)
	assert.Equal(t, 5, work.CoverId)
}

func TestSQLiteCatalogRepository_ListUnembeddedPruneAndTops_Catalog_Works(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	db := setupDB(t)
	repo := repositories.NewSQLiteCatalogRepository(db)
	signals := repositories.NewSQLiteSignalRepository(db)
	for _, w := range []*models.Work{
		{Key: "C1", Title: "Old candidate", Origin: models.OriginCandidate},
		{Key: "C2", Title: "Signal candidate", Origin: models.OriginCandidate, AuthorKeys: []string{"A2"}, Subjects: []string{"Horror", "Fiction"}},
		{Key: "U1", Title: "User", Origin: models.OriginUser, AuthorKeys: []string{"A1"}, Subjects: []string{"Thriller", "nyt:best", "Horror"}},
	} {
		_, err := repo.UpsertWork(ctx, w)
		require.NoError(t, err)
	}
	matchedSignal(t, signals, 1, "sent:1", models.SignalBook, "U1")
	matchedSignal(t, signals, 2, "sent:2", models.SignalBook, "C2")
	matchedSignal(t, signals, 2, "wish:3", models.SignalAuthor, "A9")
	_, err := db.Exec(`UPDATE catalog_work SET lastSeenAt = '2026-01-01 00:00:00'`)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO recommendation (userId, section, rank, workKey, score) VALUES (1, 'for_you', 1, 'C1', 0.5)`)
	require.NoError(t, err)

	// Act
	unembedded, err := repo.ListUnembedded(ctx, 10)
	require.NoError(t, err)
	subjects, err := repo.TopSubjects(ctx, 5)
	require.NoError(t, err)
	authors, err := repo.TopAuthorKeys(ctx, 5)
	require.NoError(t, err)
	pruned, err := repo.PruneCandidates(ctx, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
	require.NoError(t, err)
	left, err := repo.GetWorks(ctx, []string{"C1", "C2", "U1"})
	require.NoError(t, err)
	var recs int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM recommendation`).Scan(&recs))

	// Assert
	require.Len(t, unembedded, 3)
	assert.Equal(t, "U1", unembedded[0].Key, "user works first")
	assert.Equal(t, []string{"Horror", "Thriller"}, subjects, "generic and prefixed subjects are skipped")
	assert.Equal(t, "A9", authors[0], "a wished author weighs more")
	assert.ElementsMatch(t, []string{"A9", "A1", "A2"}, authors)
	assert.Equal(t, 1, pruned)
	assert.NotContains(t, left, "C1")
	assert.Contains(t, left, "C2", "a candidate someone sent stays")
	assert.Zero(t, recs, "recommendations of pruned works go too")
}

func TestSQLiteSignalRepository_Lifecycle_TracksSignalsAndStaleUsers(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	repo := repositories.NewSQLiteSignalRepository(setupDB(t))

	// Act
	first, err := repo.AddSignal(ctx, &models.Signal{UserId: 1, Ref: "wish:1", Source: models.SignalWishlist, Kind: models.SignalBook, Title: "La nena"})
	require.NoError(t, err)
	dup, err := repo.AddSignal(ctx, &models.Signal{UserId: 1, Ref: "wish:1", Source: models.SignalWishlist, Kind: models.SignalBook, Title: "La nena"})
	require.NoError(t, err)
	_, err = repo.AddSignal(ctx, &models.Signal{UserId: 2, Ref: "wish:2", Source: models.SignalWishlist, Kind: models.SignalAuthor, Title: "Mola"})
	require.NoError(t, err)
	stale, err := repo.TakeStale(ctx)
	require.NoError(t, err)
	staleAgain, err := repo.TakeStale(ctx)
	require.NoError(t, err)

	pending, err := repo.ListPending(ctx, 10, 3)
	require.NoError(t, err)
	require.NoError(t, repo.Resolve(ctx, pending[0].Id, models.SignalMatched, "OL1W"))
	matched, err := repo.ListMatched(ctx, 1)
	require.NoError(t, err)
	users, err := repo.ListUsers(ctx)
	require.NoError(t, err)

	removedUsers, err := repo.RemoveWishlistExcept(ctx, map[string]bool{"wish:1": true})
	require.NoError(t, err)
	pendingAfter, err := repo.ListPending(ctx, 10, 3)
	require.NoError(t, err)

	require.NoError(t, repo.Dismiss(ctx, 1, "OL5W"))
	dismissed, err := repo.ListDismissed(ctx, 1)
	require.NoError(t, err)
	missing, err := repo.GetState(ctx, "cursor")
	require.NoError(t, err)
	require.NoError(t, repo.SetState(ctx, "cursor", "a"))
	require.NoError(t, repo.SetState(ctx, "cursor", "b"))
	state, err := repo.GetState(ctx, "cursor")
	require.NoError(t, err)

	// Assert
	assert.True(t, first)
	assert.False(t, dup)
	assert.ElementsMatch(t, []int{1, 2}, stale)
	assert.Empty(t, staleAgain)
	require.Len(t, matched, 1)
	assert.Equal(t, "OL1W", matched[0].MatchedKey)
	assert.Equal(t, 1, matched[0].Attempts)
	assert.Equal(t, []int{1}, users)
	assert.Equal(t, []int{2}, removedUsers)
	assert.Empty(t, pendingAfter)
	assert.Equal(t, map[string]bool{"OL5W": true}, dismissed)
	assert.Empty(t, missing)
	assert.Equal(t, "b", state)
}

func TestSQLiteRecommendationRepository_ReplaceAndList_Recommendations_JoinsWorksAndReasons(t *testing.T) {
	t.Parallel()

	// Arrange
	ctx := context.Background()
	db := setupDB(t)
	catalog := repositories.NewSQLiteCatalogRepository(db)
	signals := repositories.NewSQLiteSignalRepository(db)
	repo := repositories.NewSQLiteRecommendationRepository(db)
	for _, w := range []*models.Work{
		{Key: "OWN", Title: "La nena", Origin: models.OriginUser},
		{Key: "R1", Title: "Las madres", Authors: []string{"Carmen Mola"}, CoverId: 3, FirstPublishYear: 2022, Origin: models.OriginCandidate},
		{Key: "R2", Title: "Dismissed", Origin: models.OriginCandidate},
	} {
		_, err := catalog.UpsertWork(ctx, w)
		require.NoError(t, err)
	}
	require.NoError(t, repo.Replace(ctx, 1, []models.Recommendation{{Section: models.SectionForYou, Rank: 9, Work: models.Work{Key: "R2"}}}))

	// Act
	err := repo.Replace(ctx, 1, []models.Recommendation{
		{Section: models.SectionForYou, Rank: 1, Work: models.Work{Key: "R1"}, Score: 0.9, Reason: &models.Work{Key: "OWN"}},
		{Section: models.SectionForYou, Rank: 2, Work: models.Work{Key: "R2"}, Score: 0.8},
		{Section: models.SectionNew, Rank: 1, Work: models.Work{Key: "R1"}, Score: 0.9},
	})
	require.NoError(t, err)
	require.NoError(t, signals.Dismiss(ctx, 1, "R2"))
	forYou, err := repo.List(ctx, 1, models.SectionForYou)
	require.NoError(t, err)
	other, err := repo.List(ctx, 2, models.SectionForYou)
	require.NoError(t, err)

	// Assert
	require.Len(t, forYou, 1)
	assert.Equal(t, "Las madres", forYou[0].Work.Title)
	assert.Equal(t, []string{"Carmen Mola"}, forYou[0].Work.Authors)
	assert.Equal(t, 3, forYou[0].Work.CoverId)
	require.NotNil(t, forYou[0].Reason)
	assert.Equal(t, "La nena", forYou[0].Reason.Title)
	assert.Empty(t, other)
}
