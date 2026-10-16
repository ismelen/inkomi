package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ismelen/inkomi/back/release-service/internal/config"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/api/handlers"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/api/routes"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/github"
	"github.com/ismelen/inkomi/back/release-service/internal/infra/releasecache"
	"github.com/ismelen/inkomi/back/release-service/internal/usecases/getlatestrelease"
)

func main() {
	config.Load()

	source := releasecache.NewCachedReleaseSource(
		github.NewGitHubReleaseSource(
			config.Env.GitHubAPIURL,
			config.Env.GitHubRepo,
			config.Env.GitHubToken,
			config.Env.APKAssetName,
		),
		config.Env.CacheTTL,
	)

	api := chi.NewRouter()
	routes.SetupAppRoutes(api, handlers.NewAppHandler(getlatestrelease.NewGetLatestReleaseUC(source)))

	log.Printf("Starting at port %s", config.Env.Port)
	if err := http.ListenAndServe(":"+config.Env.Port, api); err != nil {
		log.Fatal(err)
	}
}
