package featureflags

import (
	"github.com/google/uuid"
	"github.com/tadoku/tadoku/services/tadoku-api/infra/fliptmanagement"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/errx"
)

const releaseWebV3 = "release-webv3"

var releaseWebV3Access = fliptmanagement.Segment{
	Key:         "release-webv3-access",
	Name:        "Release webv3 access",
	Description: "Account UUIDs explicitly granted access to release webv3.",
}

type PublicDecisions struct {
	ReleaseWebV3 bool
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
	if flagKey != releaseWebV3 {
		return fliptmanagement.Segment{}, errx.NewInvalidInputError("feature flag is not supported")
	}
	return releaseWebV3Access, nil
}
