package wires

import (
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/openlibrary"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/infra/repositories"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/additem"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/authorworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/backfillworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/getwork"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/listitems"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/lookup"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/removeitem"
	"github.com/ismelen/inkomi/back/api/internal/modules/wishlist/usecases/resolvework"
)

// backfillInterval spaces the retries of the books whose cover and synopsis couldn't be found.
const backfillInterval = 6 * time.Hour

func Wishlist(d *Deps, r chi.Router) {
	repo := repositories.NewSQLiteItemRepository(d.WishlistDB)
	works := repositories.NewSQLiteWorkRepository(d.WishlistDB)
	catalog := openlibrary.NewOpenLibraryCatalogLookup(d.OpenLibrary)
	resolver := resolvework.NewResolveWorkUC(repo, works, catalog)
	backfillworks.NewBackfillWorksUC(repo, resolver).Start(d.Ctx, backfillInterval)
	r.Group(func(r chi.Router) {
		r.Use(d.AuthMiddleware)
		routes.SetupWishlistRoutes(r, handlers.NewWishlistHandler(
			additem.NewAddItemUC(repo, resolver),
			removeitem.NewRemoveItemUC(repo),
			listitems.NewListItemsUC(repo),
			lookup.NewLookupUC(catalog),
			authorworks.NewAuthorWorksUC(catalog),
			getwork.NewGetWorkUC(works, catalog),
		))
	})
}
