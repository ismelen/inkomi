package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/requtil"
)

func SetupBookRoutes(api chi.Router, handler *handlers.BooksHandler) {
	api.Route("/books", func(r chi.Router) {
		r.Get("/search", requtil.Wrap(handler.HandleSearch))
	})
}
