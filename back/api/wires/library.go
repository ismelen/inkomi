package wires

import (
	"log"

	"github.com/go-chi/chi/v5"

	"github.com/ismelen/inkomi/back/api/internal/config"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/infra/libgen"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/handlers"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/ui/routes"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/downloadbook"
	"github.com/ismelen/inkomi/back/api/internal/modules/library/usecases/searchbooks"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobs"
	"github.com/ismelen/inkomi/back/api/internal/platform/queue"
)

func Library(d *Deps, r chi.Router) {
	slumURL := config.Env.SlumURL
	if slumURL == "" {
		slumURL = libgen.SlumURL
	}
	library := libgen.NewLibgenLibrary(
		libgen.NewSlumMirrorList(slumURL, libgen.FallbackMirrors()),
		config.Env.MirrorRefreshCooldown,
	)
	library.Start(d.Ctx, config.Env.MirrorRefreshInterval)

	downloadBookUC := downloadbook.NewDownloadBookUC(d.Queue, library, d.Cloud)
	// network bound: one job waiting on the network doesn't need to hold back the next one
	if err := d.Queue.SubscribeWith(string(jobs.SubjectLibrary), queue.SubscribeOptions{Concurrency: 2}, jobrun.Handle(d.Queue, downloadBookUC.Execute, jobrun.WithCancellations(d.Cancels))); err != nil {
		log.Fatal(err)
	}

	r.Group(func(r chi.Router) {
		r.Use(d.AuthMiddleware)
		routes.SetupBookRoutes(r, handlers.NewBooksHandler(searchbooks.NewSearchBooksUC(library)))
	})
}
