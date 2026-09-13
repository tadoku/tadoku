package e2e_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func TestListActiveAnnouncements(t *testing.T) {
	const directory = "testdata/list_active_announcements"
	cases, err := os.ReadDir(directory)
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no announcement test cases found")
	}

	for _, testCase := range cases {
		t.Run(testCase.Name(), func(t *testing.T) {
			if !testCase.IsDir() {
				t.Fatal("expected a test case directory")
			}
			path := filepath.Join(directory, testCase.Name())
			reset(t, filepath.Join(path, "setup.sql"))
			timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
				checkHTTPGolden(t, api.handler, path)
			})

			if api.proxied.Load() != 0 {
				t.Error("native read contacted an upstream")
			}
		})
	}
}
