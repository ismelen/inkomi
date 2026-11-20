package repositories

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

const (
	sqliteTimeLayout = "2006-01-02 15:04:05"
	workColumns      = `workKey, title, authors, authorKeys, subjects, languages, firstPublishYear, coverId, description, origin`
	queryChunk       = 500
)

// genericSubjects are Open Library subjects about the edition or the library, not the book.
var genericSubjects = map[string]bool{
	"fiction": true, "accessible book": true, "protected daisy": true, "in library": true,
	"lending library": true, "large type books": true, "open library staff picks": true,
	"long now manual for civilization": true, "overdrive": true, "novela": true, "ficción": true,
	"literature": true, "general": true, "fiction, general": true, "translations into english": true,
}

type SQLiteCatalogRepository struct {
	db *sql.DB
}

var _ ports.ICatalogRepository = (*SQLiteCatalogRepository)(nil)

func NewSQLiteCatalogRepository(db *sql.DB) *SQLiteCatalogRepository {
	return &SQLiteCatalogRepository{db: db}
}

func (r *SQLiteCatalogRepository) UpsertWork(ctx context.Context, work *models.Work) (bool, error) {
	authors, authorKeys, subjects, languages := toJSON(work.Authors), toJSON(work.AuthorKeys), toJSON(work.Subjects), toJSON(work.Languages)
	year, cover := nullInt(work.FirstPublishYear), nullInt(work.CoverId)

	result, err := r.db.ExecContext(ctx, `
		INSERT INTO catalog_work (workKey, title, authors, authorKeys, subjects, languages, firstPublishYear, coverId, description, origin)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (workKey) DO NOTHING`,
		work.Key, work.Title, authors, authorKeys, subjects, languages, year, cover, work.Description, work.Origin)
	if err != nil {
		return false, err
	}
	if inserted, err := result.RowsAffected(); err != nil || inserted == 1 {
		return inserted == 1, err
	}

	// richer data wins; the embedding goes when the text it was made from changes
	_, err = r.db.ExecContext(ctx, `
		UPDATE catalog_work SET
			embedding = CASE WHEN title <> ?2
				OR json_array_length(?4) > json_array_length(subjects)
				OR (?6 <> '' AND description <> ?6) THEN NULL ELSE embedding END,
			embeddedAt = CASE WHEN title <> ?2
				OR json_array_length(?4) > json_array_length(subjects)
				OR (?6 <> '' AND description <> ?6) THEN NULL ELSE embeddedAt END,
			title = ?2,
			authors = CASE WHEN json_array_length(?3) >= json_array_length(authors) THEN ?3 ELSE authors END,
			authorKeys = CASE WHEN json_array_length(?8) >= json_array_length(authorKeys) THEN ?8 ELSE authorKeys END,
			subjects = CASE WHEN json_array_length(?4) > json_array_length(subjects) THEN ?4 ELSE subjects END,
			languages = CASE WHEN json_array_length(?5) > json_array_length(languages) THEN ?5 ELSE languages END,
			description = CASE WHEN ?6 <> '' THEN ?6 ELSE description END,
			firstPublishYear = COALESCE(?9, firstPublishYear),
			coverId = COALESCE(?10, coverId),
			origin = CASE WHEN origin = 'user' THEN 'user' ELSE ?7 END,
			lastSeenAt = CURRENT_TIMESTAMP
		WHERE workKey = ?1`,
		work.Key, work.Title, authors, subjects, languages, work.Description, work.Origin, authorKeys, year, cover)
	return false, err
}

func (r *SQLiteCatalogRepository) GetWorks(ctx context.Context, keys []string) (map[string]models.Work, error) {
	works := make(map[string]models.Work, len(keys))
	for start := 0; start < len(keys); start += queryChunk {
		chunk := keys[start:min(start+queryChunk, len(keys))]
		args := make([]any, len(chunk))
		for i, key := range chunk {
			args[i] = key
		}
		query := `SELECT ` + workColumns + ` FROM catalog_work WHERE workKey IN (?` + strings.Repeat(",?", len(chunk)-1) + `)`
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		err = scanWorks(rows, func(work models.Work) { works[work.Key] = work })
		if err != nil {
			return nil, err
		}
	}
	return works, nil
}

func (r *SQLiteCatalogRepository) ListUnembedded(ctx context.Context, limit int) ([]models.Work, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+workColumns+` FROM catalog_work
		WHERE embedding IS NULL
		ORDER BY origin = 'user' DESC, createdAt, workKey
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	var works []models.Work
	err = scanWorks(rows, func(work models.Work) { works = append(works, work) })
	return works, err
}

func (r *SQLiteCatalogRepository) SaveEmbedding(ctx context.Context, key string, embedding models.Embedding) error {
	_, err := r.db.ExecContext(ctx, `UPDATE catalog_work SET embedding = ?, embeddedAt = CURRENT_TIMESTAMP WHERE workKey = ?`,
		encodeEmbedding(embedding), key)
	return err
}

func (r *SQLiteCatalogRepository) ListVectors(ctx context.Context) ([]models.WorkVector, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT workKey, title, authorKeys, languages, firstPublishYear, embedding
		FROM catalog_work WHERE embedding IS NOT NULL`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vectors []models.WorkVector
	for rows.Next() {
		var v models.WorkVector
		var authorKeys, languages string
		var year sql.NullInt64
		var blob []byte
		if err := rows.Scan(&v.Key, &v.Title, &authorKeys, &languages, &year, &blob); err != nil {
			return nil, err
		}
		v.AuthorKeys, v.Languages, v.FirstPublishYear = fromJSON(authorKeys), fromJSON(languages), int(year.Int64)
		if v.Embedding, err = decodeEmbedding(blob); err != nil {
			return nil, fmt.Errorf("embedding of %s: %w", v.Key, err)
		}
		vectors = append(vectors, v)
	}
	return vectors, rows.Err()
}

func (r *SQLiteCatalogRepository) PruneCandidates(ctx context.Context, before time.Time) (int, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM catalog_work
		WHERE origin = 'candidate' AND lastSeenAt < ?
			AND workKey NOT IN (SELECT matchedKey FROM user_signal WHERE kind = 'book' AND status = 'matched')`,
		before.UTC().Format(sqliteTimeLayout))
	if err != nil {
		return 0, err
	}
	pruned, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if pruned > 0 {
		if _, err := r.db.ExecContext(ctx, `DELETE FROM recommendation WHERE workKey NOT IN (SELECT workKey FROM catalog_work)`); err != nil {
			return 0, err
		}
	}
	return int(pruned), nil
}

func (r *SQLiteCatalogRepository) TopSubjects(ctx context.Context, limit int) ([]string, error) {
	counts := map[string]int{}
	names := map[string]string{}
	err := r.eachUserWork(ctx, `c.subjects`, func(value string) {
		for _, subject := range fromJSON(value) {
			key := strings.ToLower(strings.TrimSpace(subject))
			if key == "" || genericSubjects[key] || strings.Contains(key, ":") || len(key) > 40 {
				continue
			}
			counts[key]++
			names[key] = subject
		}
	})
	if err != nil {
		return nil, err
	}
	top := topKeys(counts, limit)
	for i, key := range top {
		top[i] = names[key]
	}
	return top, nil
}

func (r *SQLiteCatalogRepository) TopAuthorKeys(ctx context.Context, limit int) ([]string, error) {
	counts := map[string]int{}
	err := r.eachUserWork(ctx, `c.authorKeys`, func(value string) {
		for _, key := range fromJSON(value) {
			counts[key]++
		}
	})
	if err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT matchedKey FROM user_signal WHERE kind = 'author' AND status = 'matched'`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		// a wished author counts as much as a few books
		counts[key] += 3
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return topKeys(counts, limit), nil
}

// eachUserWork calls f with column of every work some user sent or wished (once per signal).
func (r *SQLiteCatalogRepository) eachUserWork(ctx context.Context, column string, f func(string)) error {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+column+` FROM catalog_work c
		JOIN user_signal s ON s.matchedKey = c.workKey AND s.kind = 'book' AND s.status = 'matched'`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return err
		}
		f(value)
	}
	return rows.Err()
}

func scanWorks(rows *sql.Rows, f func(models.Work)) error {
	defer rows.Close()
	for rows.Next() {
		var work models.Work
		var authors, authorKeys, subjects, languages string
		var year, cover sql.NullInt64
		if err := rows.Scan(&work.Key, &work.Title, &authors, &authorKeys, &subjects, &languages, &year, &cover, &work.Description, &work.Origin); err != nil {
			return err
		}
		work.Authors, work.AuthorKeys, work.Subjects, work.Languages = fromJSON(authors), fromJSON(authorKeys), fromJSON(subjects), fromJSON(languages)
		work.FirstPublishYear, work.CoverId = int(year.Int64), int(cover.Int64)
		f(work)
	}
	return rows.Err()
}

func topKeys(counts map[string]int, limit int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if counts[keys[i]] != counts[keys[j]] {
			return counts[keys[i]] > counts[keys[j]]
		}
		return keys[i] < keys[j]
	})
	return keys[:min(limit, len(keys))]
}

func toJSON(values []string) string {
	if len(values) == 0 {
		return "[]"
	}
	data, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func fromJSON(value string) []string {
	var values []string
	if err := json.Unmarshal([]byte(value), &values); err != nil {
		return nil
	}
	return values
}

func nullInt(value int) sql.NullInt64 {
	return sql.NullInt64{Int64: int64(value), Valid: value > 0}
}
