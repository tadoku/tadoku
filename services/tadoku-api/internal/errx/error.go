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
	kind    Kind
	message string
	cause   error
}

func NewInvalidInputError(message string) *Error {
	return &Error{kind: InvalidInput, message: message}
}

func NewUnauthorizedError(message string) *Error {
	return &Error{kind: Unauthorized, message: message}
}

func NewForbiddenError(message string) *Error {
	return &Error{kind: Forbidden, message: message}
}

func NewNotFoundError(message string) *Error {
	return &Error{kind: NotFound, message: message}
}

// NewUnavailableError preserves the underlying cause, which may be nil.
func NewUnavailableError(message string, cause error) *Error {
	return &Error{kind: Unavailable, message: message, cause: cause}
}

func (e *Error) Error() string {
	if e.cause == nil {
		return e.message
	}
	if e.message == "" {
		return e.cause.Error()
	}
	return e.message + ": " + e.cause.Error()
}

func (e *Error) Unwrap() error {
	return e.cause
}

// KindOf reads the outermost application error's metadata, including through
// ordinary error wrapping. Unclassified errors and nil return Unknown.
func KindOf(err error) Kind {
	var applicationError *Error
	if errors.As(err, &applicationError) && applicationError != nil {
		return applicationError.kind
	}
	return Unknown
}
