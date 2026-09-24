package featureflags

import (
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

const releaseLogEntryV2 = "release-log-entry-v2"

var releaseLogEntryV2Access = fliptmanagement.Segment{
	Key:         "release-log-entry-v2-access",
	Name:        "Release log entry v2 access",
	Description: "Kratos UUIDs explicitly granted access to release log entry v2.",
}

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

func validate(flagKey string, targetUserID uuid.UUID) (fliptmanagement.Segment, error) {
	if targetUserID == uuid.Nil {
		return fliptmanagement.Segment{}, errx.NewInvalidInputError("target user ID must not be nil")
	}
	if flagKey != releaseLogEntryV2 {
		return fliptmanagement.Segment{}, errx.NewInvalidInputError("feature flag is not supported")
	}
	return releaseLogEntryV2Access, nil
}
