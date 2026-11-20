package repositories

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

type SQLiteRecommendationRepository struct {
	db *sql.DB
}

var _ ports.IRecommendationRepository = (*SQLiteRecommendationRepository)(nil)

func NewSQLiteRecommendationRepository(db *sql.DB) *SQLiteRecommendationRepository {
	return &SQLiteRecommendationRepository{db: db}
}

func (r *SQLiteRecommendationRepository) Replace(ctx context.Context, userId int, recs []models.Recommendation) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, tx.Rollback())
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM recommendation WHERE userId = ?`, userId); err != nil {
		return err
	}
	for _, rec := range recs {
		reason := ""
		if rec.Reason != nil {
			reason = rec.Reason.Key
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO recommendation (userId, section, rank, workKey, score, reasonWorkKey)
			VALUES (?, ?, ?, ?, ?, ?)`, userId, rec.Section, rec.Rank, rec.Work.Key, rec.Score, reason); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *SQLiteRecommendationRepository) List(ctx context.Context, userId int, section models.Section) ([]models.Recommendation, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT rec.rank, rec.score, rec.computedAt,
			w.workKey, w.title, w.authors, w.authorKeys, w.subjects, w.languages, w.firstPublishYear, w.coverId, w.description, w.origin,
			COALESCE(reason.workKey, ''), COALESCE(reason.title, '')
		FROM recommendation rec
		JOIN catalog_work w ON w.workKey = rec.workKey
		LEFT JOIN catalog_work reason ON reason.workKey = rec.reasonWorkKey
		WHERE rec.userId = ? AND rec.section = ?
			AND rec.workKey NOT IN (SELECT workKey FROM dismissed_work WHERE userId = rec.userId)
		ORDER BY rec.rank`, userId, section)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recs := []models.Recommendation{}
	for rows.Next() {
		rec := models.Recommendation{Section: section}
		w := &rec.Work
		var authors, authorKeys, subjects, languages, reasonKey, reasonTitle string
		var year, cover sql.NullInt64
		if err := rows.Scan(&rec.Rank, &rec.Score, &rec.ComputedAt,
			&w.Key, &w.Title, &authors, &authorKeys, &subjects, &languages, &year, &cover, &w.Description, &w.Origin,
			&reasonKey, &reasonTitle); err != nil {
			return nil, err
		}
		w.Authors, w.AuthorKeys, w.Subjects, w.Languages = fromJSON(authors), fromJSON(authorKeys), fromJSON(subjects), fromJSON(languages)
		w.FirstPublishYear, w.CoverId = int(year.Int64), int(cover.Int64)
		if reasonKey != "" {
			rec.Reason = &models.Work{Key: reasonKey, Title: reasonTitle}
		}
		recs = append(recs, rec)
	}
	return recs, rows.Err()
}
