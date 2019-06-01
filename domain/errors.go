package domain

import (
	"database/sql"

	"github.com/srvc/fail"
)

// ErrNotFound for when an entity could not be found
var ErrNotFound = sql.ErrNoRows

// ErrInsufficientPermissions for when access to a resource is denied
var ErrInsufficientPermissions = fail.New("need higher permissions for this resource")

// WrapError wraps errors except for domain logic related ones
func WrapError(err error, annotators ...fail.Annotator) error {
	if err == ErrNotFound {
		return err
	}
	if err == ErrInsufficientPermissions {
		return err
	}

	return fail.Wrap(err, annotators...)
}
