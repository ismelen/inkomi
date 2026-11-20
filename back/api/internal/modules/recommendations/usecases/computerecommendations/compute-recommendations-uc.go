package computerecommendations

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/ports"
)

const (
	perSection = 30
	// wishedAuthorWorks is how many works of a wished author stand for it in the profile.
	wishedAuthorWorks  = 5
	wishedAuthorWeight = 0.7
	languageBoost      = 0.04
	// sameTitle is the word overlap above which a candidate is another edition of a user's work.
	sameTitle         = 0.8
	wishedAuthorBoost = 0.05
	newYears          = 1
)

// ComputeRecommendationsUC ranks the catalog for users: their profile is the weighted mean of
// the embeddings of what they sent and wished, and every embedded work is scored against it.
type ComputeRecommendationsUC struct {
	catalog ports.ICatalogRepository
	signals ports.ISignalRepository
	recs    ports.IRecommendationRepository
	// languages are the Open Library language codes (e.g. "spa") the users read in: works with
	// an edition in them rank a bit higher.
	languages map[string]bool
	now       func() time.Time
}

func NewComputeRecommendationsUC(catalog ports.ICatalogRepository, signals ports.ISignalRepository, recs ports.IRecommendationRepository, languages []string, now func() time.Time) *ComputeRecommendationsUC {
	set := make(map[string]bool, len(languages))
	for _, language := range languages {
		set[language] = true
	}
	return &ComputeRecommendationsUC{catalog: catalog, signals: signals, recs: recs, languages: set, now: now}
}

// Execute computes and stores the recommendations of every user in userIds.
func (c *ComputeRecommendationsUC) Execute(ctx context.Context, userIds []int) error {
	if len(userIds) == 0 {
		return nil
	}
	vectors, err := c.catalog.ListVectors(ctx)
	if err != nil {
		return err
	}
	for _, userId := range userIds {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		recs, err := c.rank(ctx, userId, vectors)
		if err != nil {
			return fmt.Errorf("rank user %d: %w", userId, err)
		}
		if err := c.recs.Replace(ctx, userId, recs); err != nil {
			return err
		}
	}
	return nil
}

type taste struct {
	ownKeys       map[string]bool
	knownAuthors  map[string]bool
	wishedAuthors map[string]bool
	// vectors and weights make the profile; own are only the user's works, for the reasons
	vectors []models.WorkVector
	weights []float64
	own     []models.WorkVector
}

func (c *ComputeRecommendationsUC) rank(ctx context.Context, userId int, vectors []models.WorkVector) ([]models.Recommendation, error) {
	signals, err := c.signals.ListMatched(ctx, userId)
	if err != nil {
		return nil, err
	}
	dismissed, err := c.signals.ListDismissed(ctx, userId)
	if err != nil {
		return nil, err
	}
	t := buildTaste(signals, vectors)
	if len(t.vectors) == 0 {
		return nil, nil
	}
	embeddings := make([]models.Embedding, len(t.vectors))
	for i := range t.vectors {
		embeddings[i] = t.vectors[i].Embedding
	}
	profile := models.WeightedMean(embeddings, t.weights)

	type scored struct {
		vector models.WorkVector
		score  float64
	}
	ownTitles := make([][]string, len(t.own))
	for i, own := range t.own {
		ownTitles[i] = models.TitleWords(own.Title)
	}

	var candidates []scored
	for _, v := range vectors {
		if t.ownKeys[v.Key] || dismissed[v.Key] || isEditionOf(v.Title, ownTitles) {
			continue
		}
		score := profile.Cosine(v.Embedding)
		if hasAny(v.Languages, c.languages) {
			score += languageBoost
		}
		candidates = append(candidates, scored{vector: v, score: score})
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].score > candidates[j].score })

	sinceYear := c.now().Year() - newYears
	var recs []models.Recommendation
	pick := func(section models.Section, maxPerAuthor int, keep func(scored) (float64, bool)) {
		var chosen []scored
		for _, candidate := range candidates {
			score, ok := keep(candidate)
			if ok {
				chosen = append(chosen, scored{vector: candidate.vector, score: score})
			}
		}
		sort.SliceStable(chosen, func(i, j int) bool { return chosen[i].score > chosen[j].score })

		picked := map[string]bool{}
		perAuthor := map[string]int{}
		for _, candidate := range chosen {
			if len(picked) == perSection {
				break
			}
			if tooMany(candidate.vector.AuthorKeys, perAuthor, maxPerAuthor) {
				continue
			}
			for _, key := range candidate.vector.AuthorKeys {
				perAuthor[key]++
			}
			picked[candidate.vector.Key] = true
			recs = append(recs, models.Recommendation{
				Section: section,
				Rank:    len(picked),
				Work:    models.Work{Key: candidate.vector.Key},
				Score:   candidate.score,
				Reason:  reason(candidate.vector, t.own),
			})
		}
	}

	pick(models.SectionForYou, 2, func(s scored) (float64, bool) { return s.score, true })
	pick(models.SectionNew, 2, func(s scored) (float64, bool) {
		if s.vector.FirstPublishYear < sinceYear {
			return 0, false
		}
		if hasAny(s.vector.AuthorKeys, t.wishedAuthors) {
			return s.score + wishedAuthorBoost, true
		}
		return s.score, true
	})
	pick(models.SectionDiscover, 1, func(s scored) (float64, bool) {
		return s.score, !hasAny(s.vector.AuthorKeys, t.knownAuthors)
	})
	return recs, nil
}

func buildTaste(signals []models.Signal, vectors []models.WorkVector) taste {
	t := taste{
		ownKeys:       map[string]bool{},
		knownAuthors:  map[string]bool{},
		wishedAuthors: map[string]bool{},
	}
	weights := map[string]float64{}
	for _, signal := range signals {
		switch signal.Kind {
		case models.SignalBook:
			t.ownKeys[signal.MatchedKey] = true
			weights[signal.MatchedKey] += signal.Weight()
		case models.SignalAuthor:
			t.wishedAuthors[signal.MatchedKey] = true
			t.knownAuthors[signal.MatchedKey] = true
		}
	}

	byAuthor := map[string]int{}
	for _, v := range vectors {
		if weight, ok := weights[v.Key]; ok {
			t.vectors = append(t.vectors, v)
			t.weights = append(t.weights, weight)
			t.own = append(t.own, v)
			for _, key := range v.AuthorKeys {
				t.knownAuthors[key] = true
			}
			continue
		}
		for _, key := range v.AuthorKeys {
			if t.wishedAuthors[key] && byAuthor[key] < wishedAuthorWorks {
				byAuthor[key]++
				t.vectors = append(t.vectors, v)
				t.weights = append(t.weights, wishedAuthorWeight)
				break
			}
		}
	}
	return t
}

// reason is the user's work closest to v.
func reason(v models.WorkVector, own []models.WorkVector) *models.Work {
	best, bestScore := "", -2.0
	for _, o := range own {
		if o.Key == v.Key {
			continue
		}
		if score := v.Embedding.Cosine(o.Embedding); score > bestScore {
			best, bestScore = o.Key, score
		}
	}
	if best == "" {
		return nil
	}
	return &models.Work{Key: best}
}

// isEditionOf reports whether title is another edition or translation of one of the user's
// titles ("Novia gitana (Edición limitada)").
func isEditionOf(title string, own [][]string) bool {
	words := models.TitleWords(title)
	for _, ownWords := range own {
		if len(ownWords) > 0 && models.Overlap(ownWords, words) >= sameTitle {
			return true
		}
	}
	return false
}

func tooMany(authors []string, counts map[string]int, limit int) bool {
	for _, key := range authors {
		if counts[key] >= limit {
			return true
		}
	}
	return false
}

func hasAny(values []string, set map[string]bool) bool {
	for _, value := range values {
		if set[value] {
			return true
		}
	}
	return false
}
