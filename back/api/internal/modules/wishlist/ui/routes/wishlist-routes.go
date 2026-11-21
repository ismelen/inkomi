package routes

import (
	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/requtil"
)

func SetupWishlistRoutes(api chi.Router, handler *handlers.WishlistHandler) {
	api.Route("/wishlist", func(r chi.Router) {
		r.Get("/", requtil.Wrap(handler.HandleList))
		r.Post("/", requtil.Wrap(handler.HandleAdd))
		r.Get("/lookup", requtil.Wrap(handler.HandleLookup))
		r.Get("/works", requtil.Wrap(handler.HandleAuthorWorks))
		r.Get("/works/{workKey}", requtil.Wrap(handler.HandleWork))
		r.Delete("/{id}", requtil.Wrap(handler.HandleRemove))
	})
}
