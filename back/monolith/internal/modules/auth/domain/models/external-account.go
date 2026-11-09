package models

// ExternalAccount is an account of an external provider whose token was verified.
type ExternalAccount struct {
	Provider      Provider
	Subject       string
	Email         string
	EmailVerified bool
}
