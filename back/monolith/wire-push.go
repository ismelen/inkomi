package main

import (
	"log"
	"net/http"

	"github.com/ismelen/inkomi/back/monolith/internal/config"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/infra/dropbox"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/infra/push"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/infra/usersettings"
	"github.com/ismelen/inkomi/back/monolith/internal/modules/push/usecases/senddone"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/monolith/internal/platform/jobs"
)

func wirePush(d *deps, auth *authAPI) {
	notifier, err := push.NewFirebasePushNotifier(d.ctx, config.Env.FirebaseCredentialsFile)
	if err != nil {
		log.Fatal(err)
	}

	// Each request sends at most one chunk, so a stalled upload fails before NATS redelivers the job.
	cloudDrive := dropbox.NewDropboxCloudDrive(&http.Client{Timeout: config.Env.NATSAckWait / 2}, dropbox.ContentURL, config.Env.DropboxChunkSize)
	userSettings := usersettings.NewAuthUserSettings(auth.getDropboxAccess, auth.getFCMToken)

	sendDoneUC := senddone.NewSendDoneUC(d.queue, d.cloud, userSettings, cloudDrive, notifier)
	if err := d.queue.Subscribe(string(jobs.SubjectDone), jobrun.Handle(d.queue, sendDoneUC.Execute)); err != nil {
		log.Fatal(err)
	}
}
