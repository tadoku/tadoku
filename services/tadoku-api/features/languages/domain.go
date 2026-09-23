// Package languages owns immersion languages and their persistence.
package languages

import (
	domainlanguages "github.com/tadoku/tadoku/services/tadoku-api/domain/languages"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

type Language = domainlanguages.Language

var (
	ErrInvalidLanguage       = errx.NewInvalidInputError("invalid language")
	ErrLanguageAlreadyExists = errx.NewConflictError("language already exists")
	ErrLanguageNotFound      = errx.NewNotFoundError("language not found")
)

type CreateLanguageParameters struct {
	Code string
	Name string
}

func (p CreateLanguageParameters) Validate() error {
	if len(p.Code) < 1 || len(p.Code) > 10 || len(p.Name) < 1 || len(p.Name) > 100 {
		return ErrInvalidLanguage
	}
	return nil
}

type UpdateLanguageParameters struct {
	Code string
	Name string
}

func (p UpdateLanguageParameters) Validate() error {
	if len(p.Name) < 1 || len(p.Name) > 100 {
		return ErrInvalidLanguage
	}
	return nil
}
