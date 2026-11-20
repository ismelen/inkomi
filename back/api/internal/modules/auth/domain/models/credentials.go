package models

import "unicode/utf8"

const (
	minPasswordLength = 8
	// maxPasswordLength bounds the work of hashing an attacker-provided password.
	maxPasswordLength = 128
	maxDeviceIdLength = 128
)

func ValidatePassword(password string) error {
	n := utf8.RuneCountInString(password)
	if n < minPasswordLength || len(password) > maxPasswordLength {
		return ErrWeakPassword
	}
	return nil
}

func ValidateDeviceId(deviceId string) error {
	if deviceId == "" || len(deviceId) > maxDeviceIdLength {
		return ErrInvalidDeviceId
	}
	return nil
}
