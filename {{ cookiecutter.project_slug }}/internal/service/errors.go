package service

import "errors"

// Sentinel errors.
var (
	ErrNotFound      = errors.New("not found")
	ErrInvalid       = errors.New("invalid input")
	ErrInvalidAPIKey = errors.New("invalid api key")
)
