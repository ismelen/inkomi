package dtos

import (
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/platform/openlibrary"
)

const maxSubjects = 5

type RecommendationResponse struct {
	WorkKey  string          `json:"workKey"`
	Title    string          `json:"title"`
	Authors  []string        `json:"authors"`
	Year     int             `json:"year"`
	CoverURL string          `json:"coverUrl"`
	Subjects []string        `json:"subjects"`
	Reason   *ReasonResponse `json:"reason"`
}

type ReasonResponse struct {
	WorkKey string `json:"workKey"`
	Title   string `json:"title"`
}

func NewRecommendationResponse(rec models.Recommendation) RecommendationResponse {
	response := RecommendationResponse{
		WorkKey:  rec.Work.Key,
		Title:    rec.Work.Title,
		Authors:  nonNil(rec.Work.Authors),
		Year:     rec.Work.FirstPublishYear,
		CoverURL: openlibrary.CoverURL(rec.Work.CoverId),
		Subjects: nonNil(rec.Work.Subjects[:min(len(rec.Work.Subjects), maxSubjects)]),
	}
	if rec.Reason != nil && rec.Reason.Title != "" {
		response.Reason = &ReasonResponse{WorkKey: rec.Reason.Key, Title: rec.Reason.Title}
	}
	return response
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
