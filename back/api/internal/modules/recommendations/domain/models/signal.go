package models

type SignalSource string

const (
	SignalSent     SignalSource = "sent"
	SignalWishlist SignalSource = "wishlist"
)

type SignalKind string

const (
	SignalBook   SignalKind = "book"
	SignalAuthor SignalKind = "author"
)

type SignalStatus string

const (
	SignalPending   SignalStatus = "pending"
	SignalMatched   SignalStatus = "matched"
	SignalUnmatched SignalStatus = "unmatched"
)

// Signal is something a user sent or wished. Ref identifies it within the user ("sent:<source
// id>", "wish:<item id>"); MatchedKey is its work (book) or author key once matched.
type Signal struct {
	Id         int64
	UserId     int
	Ref        string
	Source     SignalSource
	Kind       SignalKind
	Title      string
	Author     string
	Status     SignalStatus
	Attempts   int
	MatchedKey string
}

// Weight is how much the signal pulls the user's taste: a wish says more than a send.
func (s *Signal) Weight() float64 {
	if s.Source == SignalWishlist {
		return 1.5
	}
	return 1
}
