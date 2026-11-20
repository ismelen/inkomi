package wires

import (
	"log"

	"github.com/ismelen/inkomi/back/api/internal/config"
	"github.com/ismelen/inkomi/back/api/internal/modules/epub/infra/epub"
	"github.com/ismelen/inkomi/back/api/internal/modules/epub/infra/kepub"
	"github.com/ismelen/inkomi/back/api/internal/modules/epub/usecases/joinepubs"
	"github.com/ismelen/inkomi/back/api/internal/modules/epub/usecases/kepubify"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobs"
	"github.com/ismelen/inkomi/back/api/internal/platform/queue"
)

func Epub(d *Deps) {
	joinEpubsUC := joinepubs.NewJoinEpubsUC(d.Queue, d.JoinStore, d.Cloud, epub.NewEpubMerger(), config.Env.JoinMaxBlockSize)
	kepubifyUC := kepubify.NewKepubifyUC(d.Queue, d.Cloud, kepub.NewKepubifyKepubifier())

	for subject, handler := range map[jobs.Subject]jobrun.Handler{
		jobs.SubjectJoin:  joinEpubsUC.Execute,
		jobs.SubjectKepub: kepubifyUC.Execute,
	} {
		opts := queue.SubscribeOptions{Concurrency: 1, Ready: d.Ready}
		if err := d.Queue.SubscribeWith(string(subject), opts, jobrun.Handle(d.Queue, handler, jobrun.WithCancellations(d.Cancels))); err != nil {
			log.Fatal(err)
		}
	}
}
