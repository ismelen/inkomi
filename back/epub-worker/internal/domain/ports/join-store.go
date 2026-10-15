package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/contracts/jobs"
)

type JoinStore interface {
	// Add stores msg under parentId and, once total messages are stored and the group is not
	// already claimed, claims it and returns all of them. It returns nil otherwise.
	Add(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error)
	// Clear removes a processed group and its claim.
	Clear(ctx context.Context, parentId string) error
	// Release drops the claim so a redelivered message can process the group again.
	Release(ctx context.Context, parentId string) error
}
