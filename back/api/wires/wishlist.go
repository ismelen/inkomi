package wires

import (
	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/openlibrary"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/repositories"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/additem"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/authorworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/listitems"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/lookup"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/removeitem"
)

func Wishlist(d *Deps, r chi.Router) {
	repo := repositories.NewSQLiteItemRepository(d.WishlistDB)
	catalog := openlibrary.NewOpenLibraryCatalogLookup(d.OpenLibrary)
	r.Group(func(r chi.Router) {
		r.Use(d.AuthMiddleware)
		routes.SetupWishlistRoutes(r, handlers.NewWishlistHandler(
			additem.NewAddItemUC(repo),
			removeitem.NewRemoveItemUC(repo),
			listitems.NewListItemsUC(repo),
			lookup.NewLookupUC(catalog),
			authorworks.NewAuthorWorksUC(catalog),
		))
	})
}
