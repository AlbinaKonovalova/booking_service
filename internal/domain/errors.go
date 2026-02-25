package domain

import "errors"

// Resource errors
var (
	ErrResourceNameEmpty      = errors.New("resource name must not be empty")
	ErrResourceNotFound       = errors.New("resource not found")
	ErrResourceAlreadyRemoved = errors.New("resource already removed")
)
