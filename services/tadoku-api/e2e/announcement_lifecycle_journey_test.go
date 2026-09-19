package e2e_test

import (
	"net/http"
	"testing"
	"time"
)

func TestAnnouncementLifecycleJourney(t *testing.T) {
	afterExpiry := fixtureInstant.Add(8 * 24 * time.Hour)

	runJourney(t, api, "AnnouncementLifecycle", []step{
		{
			request: "nothing_announced",
			as:      guest,
			want:    http.StatusOK,
			others:  cast{none: http.StatusBadRequest, reader: http.StatusOK, banned: http.StatusForbidden},
		},
		{
			request: "create_announcement",
			as:      admin,
			want:    http.StatusCreated,
			others: cast{
				none:    http.StatusBadRequest,
				guest:   http.StatusUnauthorized,
				reader:  http.StatusForbidden,
				reader2: http.StatusForbidden,
				banned:  http.StatusForbidden,
			},
		},
		{
			request: "announced",
			as:      guest,
			want:    http.StatusOK,
			others:  cast{reader2: http.StatusOK},
		},
		{
			request: "expired",
			as:      guest,
			want:    http.StatusOK,
			at:      afterExpiry,
			others:  cast{reader2: http.StatusOK},
		},
		{
			request: "listed_for_admin",
			as:      admin,
			want:    http.StatusOK,
			at:      afterExpiry,
			others:  cast{guest: http.StatusUnauthorized, reader: http.StatusForbidden},
		},
		{
			request: "delete_announcement",
			as:      admin,
			want:    http.StatusNoContent,
			at:      afterExpiry,
			others:  cast{guest: http.StatusUnauthorized, reader: http.StatusForbidden, banned: http.StatusForbidden},
		},
		{
			request: "gone",
			as:      admin,
			want:    http.StatusNotFound,
			at:      afterExpiry,
			others:  cast{reader: http.StatusForbidden},
		},
		{verify: "soft_deleted"},
	})
}
