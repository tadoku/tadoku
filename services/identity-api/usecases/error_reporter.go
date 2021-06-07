//go:generate gex mockgen -source=error_reporter.go -package usecases -destination=error_reporter_mock.go

package usecases

// ErrorReporter sends off errors for later inspection
type ErrorReporter interface {
	Capture(err error)
}
