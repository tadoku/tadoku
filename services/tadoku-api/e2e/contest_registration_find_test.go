package e2e_test

import (
	"net/http"
	"testing"
)

func TestFindContestRegistration(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"existing", "registration", "sorts", "languages"}, want: http.StatusOK},
		{description: []string{"deleted", "contest", "registration", "remains", "visible"}, want: http.StatusOK},
		{description: []string{"other", "user", "has", "no", "registration"}, want: http.StatusNoContent},
		{description: []string{"orphan", "registration", "is", "missing"}, want: http.StatusNoContent},
		{
			description: []string{"invalid", "contest", "id"},
			want:        http.StatusBadRequest,
		},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"invalid", "signed", "subject"}, want: http.StatusInternalServerError},
	}

	for _, test := range tests {
		name := APITestName("FindContestRegistration", test.want, test.description...)

		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
