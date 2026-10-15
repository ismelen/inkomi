package main

import (
	"context"
	"log"
	"net/http"

	"github.com/ismelen/inkomi/back/common/cloud"
	"github.com/ismelen/inkomi/back/common/httpauth"
	"github.com/ismelen/inkomi/back/common/jobrun"
	"github.com/ismelen/inkomi/back/common/jobs"
	"github.com/ismelen/inkomi/back/common/queue"
	"github.com/ismelen/inkomi/back/push-service/internal/config"
	"github.com/ismelen/inkomi/back/push-service/internal/infra/authclient"
	"github.com/ismelen/inkomi/back/push-service/internal/infra/dropbox"
	"github.com/ismelen/inkomi/back/push-service/internal/infra/push"
	"github.com/ismelen/inkomi/back/push-service/internal/usecases/senddone"
)

func main() {
	config.Load()
	ctx := context.Background()

	queue, err := queue.NewNatsQueue(config.Env.NATSUrl, config.Env.NATSAckWait, config.Env.NATSMaxDeliver)
	if err != nil {
		log.Fatal(err)
	}
	defer queue.Close()

	cloudStorage, err := cloud.NewCloudflareR2Storage(
		ctx,
		config.Env.R2AccountId,
		config.Env.R2AccessKeyId,
		config.Env.R2SecretAccessKey,
		config.Env.R2BucketName,
	)
	if err != nil {
		log.Fatal(err)
	}

	signer, err := httpauth.NewServiceTokenSigner(config.Env.ServiceName, config.Env.ServicePrivateKey, config.Env.ServiceTokenTTL)
	if err != nil {
		log.Fatal(err)
	}
	userSettings := authclient.NewHTTPUserSettings(
		&http.Client{Timeout: config.Env.AuthServiceTimeout},
		config.Env.AuthServiceURL,
		config.Env.AuthServiceAudience,
		signer,
	)

	notifier, err := push.NewFirebasePushNotifier(ctx, config.Env.FirebaseCredentialsFile)
	if err != nil {
		log.Fatal(err)
	}

	// Each request sends at most one chunk, so a stalled upload fails before NATS redelivers the job.
	cloudDrive := dropbox.NewDropboxCloudDrive(&http.Client{Timeout: config.Env.NATSAckWait / 2}, dropbox.ContentURL, config.Env.DropboxChunkSize)

	sendDoneUC := senddone.NewSendDoneUC(queue, cloudStorage, userSettings, cloudDrive, notifier)
	if err := queue.Subscribe(string(jobs.SubjectDone), jobrun.Handle(queue, sendDoneUC.Execute)); err != nil {
		log.Fatal(err)
	}

	select {}
}
