package identity

import (
	"errors"
	"testing"
)

func TestParseDisplayName(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, raw, want, wantCode string
	}{
		{"pads trimmed", "  Reader  ", "Reader", ""},
		{"one visible after trim", " a", "", "invalid_display_name"},
		{"whitespace only", "   ", "", "invalid_display_name"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDisplayName(tc.raw)
			if tc.wantCode == "" {
				if err != nil {
					t.Fatalf("ParseDisplayName(%q) error = %v, want %q", tc.raw, err, tc.want)
				}
				if got.String() != tc.want {
					t.Errorf("ParseDisplayName(%q) = %q, want %q", tc.raw, got, tc.want)
				}
				return
			}
			var identityErr *Error
			if !errors.As(err, &identityErr) || identityErr.Code != tc.wantCode {
				t.Errorf("ParseDisplayName(%q) = %v, want code %q", tc.raw, err, tc.wantCode)
			}
		})
	}
}

func TestAppHostAllowlist(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		url     string
		allowed bool
	}{
		{"https://tadoku.app", true},
		{"https://account.tadoku.app", true},
		{"https://next.tadoku.app", true},
		{"https://alice.preview.tadoku.app", true},
		{"https://alice.account.preview.tadoku.app", true},
		{"https://alice.next.preview.tadoku.app", true},
		{"https://next.tadoku.dev.lab", true},
		{"https://alice.next.tadoku.dev.lab", true},
		{"https://NEXT.tadoku.app", true},
		{"https://evil.tadoku.app", false},
		{"https://www.next.tadoku.app", false},
		{"https://next.tadoku.app.evil.test", false},
		{"https://evilnext.tadoku.app", false},
		{"https://a.b.next.preview.tadoku.app", false},
		{"https://-alice.next.preview.tadoku.app", false},
		{"https://alice.other.preview.tadoku.app", false},
		{"http://next.tadoku.app", false},
		{"https://next.tadoku.app:443", false},
		{"https://user@next.tadoku.app", false},
	} {
		originErr := ValidateOrigin(tc.url)
		if (originErr == nil) != tc.allowed {
			t.Errorf("ValidateOrigin(%q) = %v, want allowed %v", tc.url, originErr, tc.allowed)
		}

		destination := tc.url + "/logs/new?language=jpn&note=a%20b#entry"
		returnErr := ValidateReturnTo(destination)
		if (returnErr == nil) != tc.allowed {
			t.Errorf("ValidateReturnTo(%q) = %v, want allowed %v", destination, returnErr, tc.allowed)
		}
	}
}
