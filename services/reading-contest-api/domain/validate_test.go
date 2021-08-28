package domain_test

import (
	"github.com/asaskevich/govalidator"
)

// This is basic validator used in domain tests
// DO NOT USE THIS ANYWHERE ELSE

// validatable knows how to validate itself
type validatable interface {
	Validate() (valid bool, err error)
}

// validate a validatable
func validate(target validatable) (bool, error) {
	if result, err := target.Validate(); err != nil {
		return result, err
	}

	return govalidator.ValidateStruct(target)
}
