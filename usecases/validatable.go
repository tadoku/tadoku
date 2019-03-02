package usecases

// Validatable knows how to validate itself
type Validatable interface {
	Validate() (bool, error)
}
