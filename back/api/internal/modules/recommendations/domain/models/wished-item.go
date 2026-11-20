package models

// WishedItem is an item of a user's wishlist, as the wishlist module reports it.
type WishedItem struct {
	Id        int64
	UserId    int
	Kind      SignalKind
	Title     string
	Author    string
	WorkKey   string
	AuthorKey string
}
