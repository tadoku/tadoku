package infra

import (
	"github.com/getsentry/sentry-go"

	"github.com/tadoku/tadoku/services/tadoku-contest-api/domain"
	"github.com/tadoku/tadoku/services/tadoku-contest-api/usecases"
)

// NewErrorReporter creates a new error reporter that sends errors to sentry
func NewErrorReporter(dsn string) (usecases.ErrorReporter, error) {
	if dsn == "" {
		return nil, nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn: dsn,
	})

	if err != nil {
		return nil, domain.WrapError(err)
	}

	return &errorReporter{}, nil
}

type errorReporter struct {
}

func (r *errorReporter) Capture(err error) {
	_ = sentry.CaptureException(err)
}
