package dtos

import "github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"

type WorkResponse struct {
	WorkKey        string   `json:"workKey"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	CoverURL       string   `json:"coverUrl"`
	Subjects       []string `json:"subjects"`
	FirstPublished string   `json:"firstPublished"`
}

func NewWorkResponse(work models.WorkDetails) WorkResponse {
	subjects := work.Subjects
	if subjects == nil {
		subjects = []string{}
	}
	return WorkResponse{
		WorkKey:        work.WorkKey,
		Title:          work.Title,
		Description:    work.Description,
		CoverURL:       work.CoverURL,
		Subjects:       subjects,
		FirstPublished: work.FirstPublished,
	}
}
