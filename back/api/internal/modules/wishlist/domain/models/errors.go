package models

import "errors"

var (
	ErrInvalidKind   = errors.New("invalid kind")
	ErrEmptyTitle    = errors.New("empty title")
	ErrTooLong       = errors.New("title or author too long")
	ErrEmptyQuery    = errors.New("empty query")
	ErrDuplicate     = errors.New("already in the wishlist")
	ErrItemNotFound  = errors.New("wishlist item not found")
	ErrWishlistFull  = errors.New("wishlist is full")
	ErrLookupFailure = errors.New("book catalog unavailable")
	ErrInvalidSort   = errors.New("invalid sort")
	ErrInvalidPage   = errors.New("invalid page")
)
