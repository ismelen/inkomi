package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/ui/dtos"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/ui/requtil"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/dismissrecommendation"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/getrecommendations"
	"github.com/ismelen/inkomi/back/api/internal/platform/httpauth"
)

type RecommendationsHandler struct {
	getUC     *getrecommendations.GetRecommendationsUC
	dismissUC *dismissrecommendation.DismissRecommendationUC
}

func NewRecommendationsHandler(getUC *getrecommendations.GetRecommendationsUC, dismissUC *dismissrecommendation.DismissRecommendationUC) *RecommendationsHandler {
	return &RecommendationsHandler{getUC: getUC, dismissUC: dismissUC}
}

// HandleList serves GET /recommendations?section=for_you|new|discover (for_you by default).
func (h *RecommendationsHandler) HandleList(r *http.Request) (*[]dtos.RecommendationResponse, error) {
	section := models.SectionForYou
	if value := r.URL.Query().Get("section"); value != "" {
		section = models.Section(value)
	}
	recs, err := h.getUC.Execute(r.Context(), userId(r), section)
	if errors.Is(err, models.ErrInvalidSection) {
		return nil, requtil.NewError(http.StatusBadRequest, err.Error())
	}
	if err != nil {
		return nil, err
	}
	response := make([]dtos.RecommendationResponse, len(recs))
	for i, rec := range recs {
		response[i] = dtos.NewRecommendationResponse(rec)
	}
	return &response, nil
}

// HandleDismiss serves POST /recommendations/{workKey}/dismiss.
func (h *RecommendationsHandler) HandleDismiss(r *http.Request) (*struct{}, error) {
	err := h.dismissUC.Execute(r.Context(), userId(r), chi.URLParam(r, "workKey"))
	if errors.Is(err, models.ErrEmptyWorkKey) {
		return nil, requtil.NewError(http.StatusBadRequest, err.Error())
	}
	return nil, err
}

func userId(r *http.Request) int {
	claims, _ := httpauth.GetUserClaims(r.Context())
	return claims.Id
}
