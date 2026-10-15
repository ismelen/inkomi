package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/library-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/library-service/internal/infra/api/requtil"
)

func SetupBookRoutes(api chi.Router, handler *handlers.BooksHandler) {
	api.Route("/books", func(r chi.Router) {
		r.Get("/search", requtil.Wrap(handler.HandleSearch))
	})
}
