package ports

import "context"

type ISynopsisLookup interface {
	// Synopsis returns the plain-text synopsis of an edition of the work in lang, or "" when
	// there is none in that language.
	Synopsis(ctx context.Context, workKey, lang string) (string, error)
}
