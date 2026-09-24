package e2e_test

import (
	"net/http"
	"testing"
)

func TestCreateContest(t *testing.T) {
	tests := []struct {
		description []string
		want        int
	}{
		{description: []string{"member", "with", "stale", "signed", "name", "keeps", "newer", "stored", "name", "and", "duplicate", "allow", "lists"}, want: http.StatusOK},
		{description: []string{"admin", "creates", "official", "past", "contest"}, want: http.StatusOK},
		{description: []string{"member", "creates", "same", "day", "contest"}, want: http.StatusOK},
		{description: []string{"malformed", "json"}, want: http.StatusBadRequest},
		{description: []string{"empty", "body"}, want: http.StatusBadRequest},
		{description: []string{"missing", "required", "dates", "and", "activities"}, want: http.StatusBadRequest},
		{description: []string{"empty", "signed", "display", "name"}, want: http.StatusBadRequest},
		{description: []string{"three", "rune", "title"}, want: http.StatusBadRequest},
		{description: []string{"official", "private"}, want: http.StatusBadRequest},
		{description: []string{"official", "language", "limit"}, want: http.StatusBadRequest},
		{description: []string{"start", "after", "end"}, want: http.StatusBadRequest},
		{description: []string{"member", "contest", "started", "yesterday"}, want: http.StatusBadRequest},
		{description: []string{"invalid", "activity"}, want: http.StatusBadRequest},
		{description: []string{"missing", "language"}, want: http.StatusBadRequest},
		{description: []string{"guest"}, want: http.StatusUnauthorized},
		{description: []string{"non", "admin", "official"}, want: http.StatusForbidden},
		{description: []string{"yearly", "quota", "reached"}, want: http.StatusForbidden},
		{description: []string{"account", "deletion", "in", "progress"}, want: http.StatusConflict},
	}

	for _, test := range tests {
		name := APITestName("CreateContest", test.want, test.description...)
		t.Run(name, func(t *testing.T) {
			runCase(t, api, name, test.want,
				implementation{name: "tadoku-api", handler: api.handler},
			)
		})
	}
}
