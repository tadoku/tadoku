package domain

import (
	"errors"

	commondomain "github.com/tadoku/tadoku/services/common/domain"
)

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

// Announcement errors
var (
	ErrAnnouncementNotFound = errors.New("announcement not found")
	ErrInvalidAnnouncement  = errors.New("unable to validate announcement")
)

// Common errors
var (
	ErrForbidden        = commondomain.ErrForbidden
	ErrAuthzUnavailable = commondomain.ErrAuthzUnavailable
	ErrRequestInvalid   = commondomain.ErrRequestInvalid
	ErrUnauthorized     = commondomain.ErrUnauthorized
)
