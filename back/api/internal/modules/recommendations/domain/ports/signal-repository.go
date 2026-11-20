package ports

import (
	"context"

	"github.com/ismelen/inkomi/back/api/internal/modules/recommendations/domain/models"
)

type ISignalRepository interface {
	// AddSignal stores a new signal; it reports false when the user already has one with its Ref.
	AddSignal(ctx context.Context, signal *models.Signal) (bool, error)
	// RemoveWishlistExcept deletes the wishlist signals whose Ref isn't in refs and returns the
	// users that lost any.
	RemoveWishlistExcept(ctx context.Context, refs map[string]bool) ([]int, error)
	// ListPending returns up to limit signals still to match, with fewer than maxAttempts tries.
	ListPending(ctx context.Context, limit, maxAttempts int) ([]models.Signal, error)
	// Resolve stores the outcome of a match attempt and counts it.
	Resolve(ctx context.Context, id int64, status models.SignalStatus, matchedKey string) error
	// ListMatched returns the matched signals of the user.
	ListMatched(ctx context.Context, userId int) ([]models.Signal, error)
	ListUsers(ctx context.Context) ([]int, error)

	MarkStale(ctx context.Context, userIds ...int) error
	// TakeStale returns and clears the users whose recommendations must be computed again.
	TakeStale(ctx context.Context) ([]int, error)

	Dismiss(ctx context.Context, userId int, workKey string) error
	ListDismissed(ctx context.Context, userId int) (map[string]bool, error)

	GetState(ctx context.Context, name string) (string, error)
	SetState(ctx context.Context, name, value string) error
}
