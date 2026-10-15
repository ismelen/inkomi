package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/contracts/jobs"
)

type JoinStore interface {
	// Add stores msg under parentId and, once total messages are stored, returns all of them and clears the group.
	// It returns nil while the group is incomplete.
	Add(ctx context.Context, parentId string, total int, msg jobs.Msg) ([]jobs.Msg, error)
}
