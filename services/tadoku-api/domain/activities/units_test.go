package activities

import (
	"errors"
	"testing"
)

func TestResolveUnit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		activityID   int32
		key          string
		languageCode string
		wantName     string
		wantModifier float32
		wantErr      error
	}{
		{name: "reading pages", activityID: 1, key: "pages", languageCode: "eng", wantName: "Page", wantModifier: 1},
		{name: "writing pages", activityID: 3, key: "pages", languageCode: "eng", wantName: "Page", wantModifier: 10},
		{
			name:         "reading characters in Japanese",
			activityID:   1,
			key:          "characters",
			languageCode: "jpn",
			wantName:     "Character",
			wantModifier: 0.0025,
		},
		{
			name:         "reading characters in another language",
			activityID:   1,
			key:          "characters",
			languageCode: "eng",
			wantName:     "Character",
			wantModifier: 0.000833333,
		},
		{
			name:         "writing characters in Cantonese",
			activityID:   3,
			key:          "characters",
			languageCode: "yue",
			wantName:     "Character",
			wantModifier: 0.025,
		},
		{name: "activity-prefixed key", activityID: 1, key: "reading_page", languageCode: "eng", wantErr: errUnknownUnit},
		{name: "retired key", activityID: 1, key: "reading_comic_page", languageCode: "eng", wantErr: errUnknownUnit},
		{name: "unit on a timed activity", activityID: 2, key: "pages", languageCode: "eng", wantErr: errUnitNotOffered},
		{name: "unknown activity", activityID: 9, key: "pages", languageCode: "eng", wantErr: errUnitNotOffered},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			unit, err := ResolveUnit(test.activityID, test.key, test.languageCode)

			if !errors.Is(err, test.wantErr) {
				t.Fatalf("ResolveUnit() error = %v, want %v", err, test.wantErr)
			}
			if unit.Name != test.wantName || unit.Modifier != test.wantModifier {
				t.Errorf(
					"ResolveUnit() = %q with modifier %v, want %q with modifier %v",
					unit.Name,
					unit.Modifier,
					test.wantName,
					test.wantModifier,
				)
			}
		})
	}
}

func TestCanonicalStoredUnitKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		activityID int32
		key        string
		want       string
	}{
		{activityID: 1, key: "reading_page", want: "pages"},
		{activityID: 3, key: "writing_character", want: "characters"},
		{activityID: 1, key: "sentences", want: "sentences"},
		{activityID: 3, key: "reading_page", want: "reading_page"},
		{activityID: 1, key: "reading_comic_page", want: "reading_comic_page"},
	}
	for _, test := range tests {
		got := CanonicalStoredUnitKey(test.activityID, test.key)

		if got != test.want {
			t.Errorf("CanonicalStoredUnitKey(%d, %q) = %q, want %q", test.activityID, test.key, got, test.want)
		}
	}
}
