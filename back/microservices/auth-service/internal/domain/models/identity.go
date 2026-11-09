package models

// Identity links an account of an external provider to a user.
type Identity struct {
	Provider Provider
	Subject  string
	UserId   int
}
