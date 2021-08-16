package main

import (
	"net/url"
	"testing"
)

func TestRemoveSubdomain(t *testing.T) {
	var tests = []struct {
		in  string
		out string
	}{
		{"https://www.tadoku.app/", "https://tadoku.app/"},
		{"https://tadoku.app/", "https://tadoku.app/"},
		{"http://tadoku.app/", "http://tadoku.app/"},
		{"https://staging.tadoku.app/", "https://staging.tadoku.app/"},
	}

	for _, test := range tests {
		url, err := url.Parse(test.in)
		if err != nil {
			t.Errorf("test case has invalid url: %v:", test.in)
		}
		got := RemoveWWWSubdomain(url)
		if got.String() != test.out {
			t.Errorf("RemoveWWWSubdomain(%v) = %v; want %v", test.in, got, test.out)
		}
	}
}
