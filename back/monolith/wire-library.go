package main

import (
	"log"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/library/infra/api/handlers"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/library/infra/api/routes"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/library/infra/libgen"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/library/usecases/downloadbook"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/library/usecases/searchbooks"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
)

func wireLibrary(d *deps, r chi.Router) {
	slumURL := config.Env.SlumURL
	if slumURL == "" {
		slumURL = libgen.SlumURL
	}
	library := libgen.NewLibgenLibrary(
		libgen.NewSlumMirrorList(slumURL, libgen.FallbackMirrors()),
		config.Env.MirrorRefreshCooldown,
	)
	library.Start(d.ctx, config.Env.MirrorRefreshInterval)

	downloadBookUC := downloadbook.NewDownloadBookUC(d.queue, library, d.cloud)
	if err := d.queue.Subscribe(string(jobs.SubjectLibrary), jobrun.Handle(d.queue, downloadBookUC.Execute)); err != nil {
		log.Fatal(err)
	}

	r.Group(func(r chi.Router) {
		r.Use(d.authMiddleware)
		routes.SetupBookRoutes(r, handlers.NewBooksHandler(searchbooks.NewSearchBooksUC(library)))
	})
}
