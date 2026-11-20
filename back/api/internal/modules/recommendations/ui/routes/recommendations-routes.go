package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/ui/requtil"
)

func SetupRecommendationsRoutes(api chi.Router, handler *handlers.RecommendationsHandler) {
	api.Route("/recommendations", func(r chi.Router) {
		r.Get("/", requtil.Wrap(handler.HandleList))
		r.Post("/{workKey}/dismiss", requtil.Wrap(handler.HandleDismiss))
	})
}
