package dtos

import "github.com/ismelen/inkomi/back/api/internal/modules/wishlist/domain/models"

type WorksPageResponse struct {
	Works   []SuggestionResponse `json:"works"`
	Total   int                  `json:"total"`
	Page    int                  `json:"page"`
	HasMore bool                 `json:"hasMore"`
}

func NewWorksPageResponse(page models.WorksPage) WorksPageResponse {
	works := make([]SuggestionResponse, len(page.Works))
	for i, w := range page.Works {
		works[i] = NewSuggestionResponse(w)
	}
	return WorksPageResponse{Works: works, Total: page.Total, Page: page.Page, HasMore: page.HasMore}
}
