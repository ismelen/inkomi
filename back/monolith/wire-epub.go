package main

import (
	"log"
	"path/filepath"
	"time"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/infra/epub"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/infra/kepub"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/infra/store"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/usecases/joinepubs"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/epub/usecases/kepubify"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/queue"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/sqlitedb"
	"github.com/ismelen/inkomi/back/monolith/migrations"
)

func wireEpub(d *deps) func() {
	db, err := sqlitedb.Open(filepath.Join(config.Env.DataDir, "jobs.db"), migrations.Joins)
	if err != nil {
		log.Fatal(err)
	}

	joinStore := store.NewSQLiteJoinStore(db, config.Env.JoinTTL, config.Env.JoinLockTTL, time.Now)
	joinEpubsUC := joinepubs.NewJoinEpubsUC(d.queue, joinStore, d.cloud, epub.NewEpubMerger(), config.Env.JoinMaxBlockSize)
	kepubifyUC := kepubify.NewKepubifyUC(d.queue, d.cloud, kepub.NewKepubifyKepubifier())

	for subject, handler := range map[jobs.Subject]jobrun.Handler{
		jobs.SubjectJoin:  joinEpubsUC.Execute,
		jobs.SubjectKepub: kepubifyUC.Execute,
	} {
		opts := queue.SubscribeOptions{Concurrency: 1, Ready: d.memoryGate.Ready}
		if err := d.queue.SubscribeWith(string(subject), opts, jobrun.Handle(d.queue, handler)); err != nil {
			log.Fatal(err)
		}
	}

	return closeLogged("jobs.db", db.Close)
}
