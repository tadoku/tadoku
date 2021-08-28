//go:generate gex mockgen -source=validator.go -package usecases -destination=validator_mock.go

package usecases

// Validator validates entities
type Validator interface {
	ValidateStruct(s interface{}) (bool, error)
	Validate(target Validatable) (bool, error)
}
