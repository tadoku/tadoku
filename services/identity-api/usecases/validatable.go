package usecases

// Validatable knows how to validate itself
type Validatable interface {
	Validate() (valid bool, err error)
}
