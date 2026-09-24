// Package languages owns immersion languages and their persistence.
package languages

import (
	domainlanguages "github.com/tadoku/tadoku/services/tadoku-api/domain/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Language = domainlanguages.Language

var (
	ErrLanguageAlreadyExists = errx.NewConflictError("language already exists")
	ErrLanguageNotFound      = errx.NewNotFoundError("language not found")
)

type CreateLanguageParameters struct {
	Code string
	Name string
}

func (p CreateLanguageParameters) Validate() error {
	if len(p.Code) < 1 {
		return errx.NewInvalidInputError("code is required")
	}
	if len(p.Code) > 10 {
		return errx.NewInvalidInputError("code must be at most 10 bytes")
	}
	if len(p.Name) < 1 {
		return errx.NewInvalidInputError("name is required")
	}
	if len(p.Name) > 100 {
		return errx.NewInvalidInputError("name must be at most 100 bytes")
	}
	return nil
}

type UpdateLanguageParameters struct {
	Code string
	Name string
}

func (p UpdateLanguageParameters) Validate() error {
	if len(p.Name) < 1 {
		return errx.NewInvalidInputError("name is required")
	}
	if len(p.Name) > 100 {
		return errx.NewInvalidInputError("name must be at most 100 bytes")
	}
	return nil
}
