package ports

import "context"

type TxManager interface {
	ExecuteTx(ctx context.Context, fn func(ctx context.Context) error) error
}
