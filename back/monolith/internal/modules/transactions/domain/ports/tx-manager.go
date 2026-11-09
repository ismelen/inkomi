package ports

import "context"

type ITxManager interface {
	ExecuteTx(ctx context.Context, fn func(ctx context.Context) error) error
}
