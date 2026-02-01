package domain

import "errors"

// Page errors
var (
	ErrPageAlreadyExists = errors.New("page with given slug already exists")
	ErrPageNotFound      = errors.New("page not found")
	ErrInvalidPage       = errors.New("unable to validate page")
)

// Post errors
var (
	ErrPostAlreadyExists = errors.New("post with given slug already exists")
	ErrPostNotFound      = errors.New("post not found")
	ErrInvalidPost       = errors.New("unable to validate post")
)

// Common errors
var (
	ErrForbidden      = errors.New("not allowed")
	ErrRequestInvalid = errors.New("request is invalid")
)
