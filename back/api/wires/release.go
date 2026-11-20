package wires

import (
	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/config"
	"github.com/ismelen/inkomi/back/api/internal/modules/release/infra/github"
	"github.com/ismelen/inkomi/back/api/internal/modules/release/infra/releasecache"
	"github.com/ismelen/inkomi/back/api/internal/modules/release/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/release/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/release/usecases/getlatestrelease"
)

func Release(r chi.Router) {
	source := releasecache.NewCachedReleaseSource(
		github.NewGitHubReleaseSource(
			config.Env.GitHubAPIURL,
			config.Env.GitHubRepo,
			config.Env.GitHubToken,
			config.Env.APKAssetName,
		),
		config.Env.CacheTTL,
	)
	routes.SetupAppRoutes(r, handlers.NewAppHandler(getlatestrelease.NewGetLatestReleaseUC(source)))
}
