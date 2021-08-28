package domain

import "github.com/srvc/fail"

// Environment of the app
type Environment string

// Environment enum
const (
	EnvProduction  Environment = "production"
	EnvDevelopment Environment = "development"
	EnvTest        Environment = "test"
)

// AllEnvironments a collection of all valid environments
var AllEnvironments = []Environment{
	EnvProduction,
	EnvDevelopment,
	EnvTest,
}

// ErrInvalidEnvironment for when an environment is not defined in our app
var ErrInvalidEnvironment = fail.New("supplied environment is not supported")

// Validate wether or the the environment is a valid one
func (env Environment) Validate() error {
	for _, e := range AllEnvironments {
		if e == env {
			return nil
		}
	}

	return ErrInvalidEnvironment
}

// ShouldSecure indicates wether or not the environment needs to have security enforced
func (env Environment) ShouldSecure() bool {
	return env == EnvProduction
}
