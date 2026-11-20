package wires

import (
	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/openlibrary"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/repositories"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/additem"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/listallitems"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/listitems"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/lookup"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/removeitem"
)

// WishlistAPI is what the recommendations module uses of the wishlist, in process.
type WishlistAPI struct {
	ListAll *listallitems.ListAllItemsUC
	// OnChange registers f to be called with the user after any change of their wishlist.
	OnChange func(f func(userId int))
}

func Wishlist(d *Deps, r chi.Router) *WishlistAPI {
	repo := repositories.NewSQLiteItemRepository(d.DiscoverDB)
	addUC := additem.NewAddItemUC(repo)
	removeUC := removeitem.NewRemoveItemUC(repo)

	r.Group(func(r chi.Router) {
		r.Use(d.AuthMiddleware)
		routes.SetupWishlistRoutes(r, handlers.NewWishlistHandler(
			addUC,
			removeUC,
			listitems.NewListItemsUC(repo),
			lookup.NewLookupUC(openlibrary.NewOpenLibraryCatalogLookup(d.OpenLibrary)),
		))
	})

	return &WishlistAPI{
		ListAll: listallitems.NewListAllItemsUC(repo),
		OnChange: func(f func(userId int)) {
			addUC.SetOnChange(f)
			removeUC.SetOnChange(f)
		},
	}
}
