package wires

import (
	"log"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/config"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/infra/embedding"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/infra/openlibrary"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/infra/repositories"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/infra/sentbooks"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/infra/wishlistreader"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/computerecommendations"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/dismissrecommendation"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/embedworks"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/expandcatalog"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/getrecommendations"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/importsignals"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/refreshcatalog"
	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/usecases/resolvesignals"
)

// Recommendations serves the recommendations and, when enabled, keeps the catalog up to date in
// the background (Open Library + the embedding model in DATA_DIR/models).
func Recommendations(d *Deps, r chi.Router, transactions *TransactionsAPI, wishlist *WishlistAPI) {
	catalog := repositories.NewSQLiteCatalogRepository(d.DiscoverDB)
	signals := repositories.NewSQLiteSignalRepository(d.DiscoverDB)
	recs := repositories.NewSQLiteRecommendationRepository(d.DiscoverDB)

	r.Group(func(r chi.Router) {
		r.Use(d.AuthMiddleware)
		routes.SetupRecommendationsRoutes(r, handlers.NewRecommendationsHandler(
			getrecommendations.NewGetRecommendationsUC(recs),
			dismissrecommendation.NewDismissRecommendationUC(signals),
		))
	})

	if !config.Env.RecsEnabled {
		log.Print("recommendations: refresh disabled (RECS_ENABLED=false)")
		return
	}
	source := openlibrary.NewOpenLibraryCatalogSource(d.OpenLibrary)
	model := embedding.NewModelFiles(
		filepath.Join(config.Env.DataDir, "models", "embeddings"),
		config.Env.EmbeddingModelURL,
		config.Env.EmbeddingOnnxFile,
	)
	refresh := refreshcatalog.NewRefreshCatalogUC(
		importsignals.NewImportSignalsUC(
			sentbooks.NewTransactionsSentBooks(transactions.ListSentBooks),
			wishlistreader.NewWishlistReader(wishlist.ListAll),
			signals,
		),
		resolvesignals.NewResolveSignalsUC(source, catalog, signals),
		expandcatalog.NewExpandCatalogUC(source, catalog, time.Now),
		embedworks.NewEmbedWorksUC(embedding.NewHugotEmbedder(model), catalog, d.Ready),
		computerecommendations.NewComputeRecommendationsUC(catalog, signals, recs, config.Env.RecsLanguages, time.Now),
		catalog,
		signals,
		config.Env.RecsRefreshBudget,
		time.Now,
	)
	wishlist.OnChange(refresh.Trigger)
	refresh.Start(d.Ctx, config.Env.RecsRefreshInterval)
}
