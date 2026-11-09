package main

import (
	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/release/infra/api/handlers"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/release/infra/api/routes"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/release/infra/github"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/release/infra/releasecache"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/release/usecases/getlatestrelease"
)

func wireRelease(r chi.Router) {
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
