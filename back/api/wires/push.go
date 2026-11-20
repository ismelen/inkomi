package wires

import (
	"log"
	"net/http"

	"github.com/ismelen/inkomi/back/api/internal/config"
	"github.com/ismelen/inkomi/back/api/internal/modules/push/infra/dropbox"
	"github.com/ismelen/inkomi/back/api/internal/modules/push/infra/push"
	"github.com/ismelen/inkomi/back/api/internal/modules/push/infra/usersettings"
	"github.com/ismelen/inkomi/back/api/internal/modules/push/usecases/senddone"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobrun"
	"github.com/ismelen/inkomi/back/api/internal/platform/jobs"
	"github.com/ismelen/inkomi/back/api/internal/platform/queue"
)

func Push(d *Deps, auth *AuthAPI) {
	notifier, err := push.NewFirebasePushNotifier(d.Ctx, config.Env.FirebaseCredentialsFile)
	if err != nil {
		log.Fatal(err)
	}

	// Each request sends at most one chunk, so a stalled upload fails before NATS redelivers the job.
	cloudDrive := dropbox.NewDropboxCloudDrive(&http.Client{Timeout: config.Env.NATSAckWait / 2}, dropbox.ContentURL, config.Env.DropboxChunkSize)
	userSettings := usersettings.NewAuthUserSettings(auth.getDropboxAccess, auth.getFCMToken)

	sendDoneUC := senddone.NewSendDoneUC(d.Queue, d.Cloud, userSettings, cloudDrive, notifier)
	// network bound: one job waiting on the network doesn't need to hold back the next one
	if err := d.Queue.SubscribeWith(string(jobs.SubjectDone), queue.SubscribeOptions{Concurrency: 2}, jobrun.Handle(d.Queue, sendDoneUC.Execute, jobrun.WithCancellations(d.Cancels))); err != nil {
		log.Fatal(err)
	}
}
