package main

import (
	"log"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/infra/epub"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/infra/kepub"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/usecases/joinepubs"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/usecases/kepubify"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/queue"
)

func wireEpub(d *deps) {
	joinEpubsUC := joinepubs.NewJoinEpubsUC(d.queue, d.joinStore, d.cloud, epub.NewEpubMerger(), config.Env.JoinMaxBlockSize)
	kepubifyUC := kepubify.NewKepubifyUC(d.queue, d.cloud, kepub.NewKepubifyKepubifier())

	for subject, handler := range map[jobs.Subject]jobrun.Handler{
		jobs.SubjectJoin:  joinEpubsUC.Execute,
		jobs.SubjectKepub: kepubifyUC.Execute,
	} {
		opts := queue.SubscribeOptions{Concurrency: 1, Ready: d.ready}
		if err := d.queue.SubscribeWith(string(subject), opts, jobrun.Handle(d.queue, handler, jobrun.WithCancellations(d.cancels))); err != nil {
			log.Fatal(err)
		}
	}
}
