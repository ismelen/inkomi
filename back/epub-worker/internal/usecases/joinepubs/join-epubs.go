package joinepubs

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ismelen/inkomi/back/contracts/jobs"
	"github.com/ismelen/inkomi/back/epub-worker/internal/domain/ports"
)

var errMissingParent = errors.New("join job without parent or total")

type JoinEpubsUC struct {
	queue     ports.Queue
	joinStore ports.JoinStore
}

func NewJoinEpubsUC(queue ports.Queue, joinStore ports.JoinStore) *JoinEpubsUC {
	return &JoinEpubsUC{
		queue:     queue,
		joinStore: joinStore,
	}
}

func (j *JoinEpubsUC) Execute() error {
	return j.queue.Subscribe(string(jobs.SubjectJoin), func(ctx context.Context, subject string, payload []byte) error {
		return j.handle(ctx, payload)
	})
}

func (j *JoinEpubsUC) handle(ctx context.Context, payload []byte) error {
	var msg jobs.Msg
	if err := json.Unmarshal(payload, &msg); err != nil {
		return j.fail(ctx, msg, err)
	}
	if msg.Parent == nil || msg.Parent.Total <= 0 {
		return j.fail(ctx, msg, errMissingParent)
	}

	msgs, err := j.joinStore.Add(ctx, msg.Parent.Id, msg.Parent.Total, msg)
	if err != nil {
		return err
	}
	if msgs == nil {
		return nil
	}

	// TODO: download msgs
	// TODO: divide secuentialy and divide by size: sum up files until block size >= 200 Mb; probably job.Msg needs to have the index inside folder and also in transaction-manager.db
	// TODO: merge each block into one epub, upload results to cloud storage, publish kepub/done for the results including parentId. We need to take into account that 1 folder may have >1 results. Plan something.

	return nil
}

// fail reports an unprocessable job; returning the publish result acks the message so it isn't retried.
func (j *JoinEpubsUC) fail(ctx context.Context, msg jobs.Msg, err error) error {
	return j.queue.Publish(ctx, string(jobs.SubjectFailed), jobs.Msg{
		Id:       msg.Id,
		UserId:   msg.UserId,
		ErrorMsg: err.Error(),
	})
}
