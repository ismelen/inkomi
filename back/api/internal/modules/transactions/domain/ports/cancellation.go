package ports

import "context"

type ICancelRegistry interface {
	// Cancel stops the queued and running jobs of ids.
	Cancel(ctx context.Context, ids []string) error
}

type IFileCleaner interface {
	// DeleteAll removes every stored file of a source.
	DeleteAll(id string) error
}

type IJoinGroups interface {
	// Clear forgets the chapters a folder's join was waiting for.
	Clear(ctx context.Context, parentId string) error
}
