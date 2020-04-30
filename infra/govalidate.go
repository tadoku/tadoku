package infra

import (
	"github.com/asaskevich/govalidator"

	"github.com/tadoku/api/domain"
)

// ConfigureCustomValidators sets up custom struct validators
func ConfigureCustomValidators() {
	govalidator.CustomTypeTagMap.Set("environment", func(i interface{}, o interface{}) bool {
		switch env := i.(type) { // type switch on the struct field being validated
		case domain.Environment:
			err := env.Validate()
			return err == nil
		}
		return false
	})

}
