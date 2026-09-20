package e2e_test

import (
	"net/http"
	"testing"
)

func TestUpsertContestRegistration(t *testing.T) {
	tests := []struct {
		description []string
		want        int
		skipParity  string
	}{
		{description: []string{"creates", "for", "private", "ended", "contest"}, want: http.StatusOK},
		{description: []string{"updates", "existing", "registration"}, want: http.StatusOK},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{
			description: []string{"missing", "body"},
			want:        http.StatusBadRequest,
			skipParity:  "the generated decoder rejects an empty body before the legacy application validates it",
		},
		{
			description: []string{"missing", "contest", "with", "missing", "body"},
			want:        http.StatusBadRequest,
			skipParity:  "the generated decoder rejects an empty body before the legacy application returns not found",
		},
		{description: []string{"no", "languages"}, want: http.StatusBadRequest},
		{description: []string{"too", "many", "languages"}, want: http.StatusBadRequest},
		{description: []string{"duplicate", "languages"}, want: http.StatusBadRequest},
		{description: []string{"unknown", "language"}, want: http.StatusBadRequest},
		{description: []string{"language", "not", "allowed"}, want: http.StatusBadRequest},
		{description: []string{"missing", "contest", "wins", "over", "invalid", "languages"}, want: http.StatusNotFound},
		{description: []string{"deleted", "contest"}, want: http.StatusNotFound},
		{description: []string{"soft", "deleted", "registration", "is", "not", "resurrected"}, want: http.StatusInternalServerError},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"banned"}, want: http.StatusForbidden},
		{description: []string{"account", "deletion", "in", "progress"}, want: http.StatusConflict},
		{description: []string{"invalid", "signed", "subject"}, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		name := APITestName("UpsertContestRegistration", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
				implementation{
					name:    "immersion-api",
					handler: legacyImmersion.handler,
					skip:    test.skipParity,
				},
			)
		})
	}
}
