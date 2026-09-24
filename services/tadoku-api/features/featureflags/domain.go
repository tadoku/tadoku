// Package featureflags evaluates flags and manages allowlisted named-user access.
package featureflags

import (
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

const releaseLogEntryV2 = "release-log-entry-v2"

var ErrUnavailable = errx.NewUnavailableError("feature access unavailable", nil)

type PublicDecisions struct {
	ReleaseLogEntryV2 bool
}

type State struct {
	Enabled     bool
	Changed     bool
	Environment string
	Revision    string
}

func validate(flagKey string, targetUserID uuid.UUID) (fliptmanagement.FeatureFlagKey, error) {
	if targetUserID == uuid.Nil {
		return "", errx.NewInvalidInputError("target user ID must not be nil")
	}
	if flagKey != releaseLogEntryV2 {
		return "", errx.NewInvalidInputError("feature flag is not supported")
	}
	return fliptmanagement.FeatureFlagReleaseLogEntryV2, nil
}
