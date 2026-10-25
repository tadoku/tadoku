package identity

import "testing"

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
