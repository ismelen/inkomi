package routes

import (
	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/requtil"
)

func SetupBookRoutes(api chi.Router, handler *handlers.BooksHandler) {
	api.Get("/books/search", requtil.Wrap(handler.HandleSearch))
}

// SetupCoverRoutes registers the public cover route (no token: <img> tags can't send one).
func SetupCoverRoutes(api chi.Router, handler *handlers.CoversHandler) {
	api.Get("/books/cover", handler.HandleCover)
}
