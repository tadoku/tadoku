package e2e_test

import (
	"net/http"
	"testing"
)

func TestImmersionContestListEvents(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest", "history"}, want: http.StatusOK},
		{description: []string{"user", "history"}, want: http.StatusOK},
		{description: []string{"admin", "history"}, want: http.StatusOK},
		{description: []string{"zero", "cursor"}, want: http.StatusOK},
		{description: []string{"first", "page"}, want: http.StatusOK},
		{description: []string{"middle", "page"}, want: http.StatusOK},
		{description: []string{"last", "page"}, want: http.StatusOK},
		{description: []string{"caught", "up"}, want: http.StatusOK},
		{description: []string{"now"}, want: http.StatusOK},
		{description: []string{"empty"}, want: http.StatusOK},
		{description: []string{"empty", "now"}, want: http.StatusOK},
		{description: []string{"private", "guest"}, want: http.StatusOK},
		{description: []string{"tenant", "history"}, want: http.StatusOK},
		{description: []string{"empty", "cursor"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "cursor"}, want: http.StatusBadRequest},
		{description: []string{"long", "cursor"}, want: http.StatusBadRequest},
		{description: []string{"unknown", "cursor"}, want: http.StatusBadRequest},
		{description: []string{"cross", "contest", "cursor"}, want: http.StatusBadRequest},
		{description: []string{"cross", "tenant", "cursor"}, want: http.StatusBadRequest},
		{description: []string{"foreign", "canonical", "cursor"}, want: http.StatusBadRequest},
		{description: []string{"zero", "limit"}, want: http.StatusBadRequest},
		{description: []string{"negative", "limit"}, want: http.StatusBadRequest},
		{description: []string{"excess", "limit"}, want: http.StatusBadRequest},
		{description: []string{"malformed", "limit"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "contest"}, want: http.StatusBadRequest},
		{description: []string{"missing", "contest"}, want: http.StatusNotFound},
		{description: []string{"deleted", "contest"}, want: http.StatusNotFound},
		{description: []string{"banned"}, want: http.StatusForbidden},
	}
	for _, test := range tests {
		name := APITestName("ImmersionContestListEvents", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}

func TestImmersionContestFindLog(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"guest", "contest", "score"}, want: http.StatusOK},
		{description: []string{"owner", "without", "registrations"}, want: http.StatusOK},
		{description: []string{"admin", "without", "registrations"}, want: http.StatusOK},
		{description: []string{"private", "contest", "score"}, want: http.StatusOK},
		{description: []string{"tenant", "contest", "score"}, want: http.StatusOK},
		{description: []string{"withdrawn"}, want: http.StatusNotFound},
		{description: []string{"unattached"}, want: http.StatusNotFound},
		{description: []string{"deleted"}, want: http.StatusNotFound},
		{description: []string{"deleted", "admin"}, want: http.StatusNotFound},
		{description: []string{"missing", "contest"}, want: http.StatusNotFound},
		{description: []string{"deleted", "contest", "admin"}, want: http.StatusNotFound},
		{description: []string{"wrong", "contest"}, want: http.StatusNotFound},
		{description: []string{"invalid", "contest"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "log"}, want: http.StatusBadRequest},
		{description: []string{"banned"}, want: http.StatusForbidden},
	}
	for _, test := range tests {
		name := APITestName("ImmersionContestFindLog", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
