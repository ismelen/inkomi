package models

import (
	"net/mail"
	"strings"
)

const maxEmailLength = 254

// NormalizeEmail trims and lowercases raw and checks it is a bare address.
func NormalizeEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" || len(email) > maxEmailLength {
		return "", ErrInvalidEmail
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email || address.Name != "" {
		return "", ErrInvalidEmail
	}
	return email, nil
}
