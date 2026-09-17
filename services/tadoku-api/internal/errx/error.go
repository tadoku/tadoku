// Package errx carries transport-neutral application failure metadata.
package errx

import "errors"

type Kind uint8

const (
	Unknown Kind = iota
	InvalidInput
	Unauthorized
	Forbidden
	NotFound
	Unavailable
)

type Error struct {
	Kind    Kind
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	if e.Message == "" {
		return e.Cause.Error()
	}
	return e.Message + ": " + e.Cause.Error()
}

func (e *Error) Unwrap() error {
	return e.Cause
}

// KindOf reads the outermost application error's metadata, including through
// ordinary error wrapping. Unclassified errors and nil return Unknown.
func KindOf(err error) Kind {
	var applicationError *Error
	if errors.As(err, &applicationError) && applicationError != nil {
		return applicationError.Kind
	}
	return Unknown
}
