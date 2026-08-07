package errors

import "errors"

var (
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidID          = errors.New("invalid identifier")
	ErrInvalidYear        = errors.New("invalid year")
	ErrInvalidLogin       = errors.New("invalid login")
	ErrWeakPassword       = errors.New("weak password")
	ErrLoginTaken         = errors.New("login already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("authentication required")
)
