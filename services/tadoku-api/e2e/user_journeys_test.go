package e2e_test

import (
	"net/http"
	"testing"
	"time"
)

// Every important user journey in the application lives in this file. Replay a
// request as other cast members only where that adds information: rejected
// identities on mutating steps, and a second user only to observe limited
// visibility of a resource.

func TestAnnouncementLifecycleJourney(t *testing.T) {
	afterExpiry := fixtureInstant.Add(8 * 24 * time.Hour)

	runJourney(t, api, "AnnouncementLifecycle", []step{
		{
			request: "nothing_announced",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "create_announcement",
			as:      admin,
			want:    http.StatusCreated,
			others: cast{
				none:   http.StatusBadRequest,
				guest:  http.StatusUnauthorized,
				user:   http.StatusForbidden,
				banned: http.StatusForbidden,
			},
		},
		{
			request: "announced",
			as:      guest,
			want:    http.StatusOK,
		},
		{
			request: "expired",
			as:      guest,
			want:    http.StatusOK,
			at:      afterExpiry,
		},
		{
			request: "listed_for_admin",
			as:      admin,
			want:    http.StatusOK,
			at:      afterExpiry,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden},
		},
		{
			request: "delete_announcement",
			as:      admin,
			want:    http.StatusNoContent,
			at:      afterExpiry,
			others:  cast{guest: http.StatusUnauthorized, user: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "gone",
			as:      admin,
			want:    http.StatusNotFound,
			at:      afterExpiry,
		},
		{verify: "soft_deleted"},
	})
}
