package e2e_test

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/timex"
)

func APITestName(operation string, status int, description ...string) string {
	return fmt.Sprintf("%s/%d_%s", operation, status, strings.Join(description, "_"))
}

func checkCaseGolden(t *testing.T, handler http.Handler, directory string, wantStatus int) {
	t.Helper()
	resetCase(t, directory)
	timex.TheWorld(time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC), func() {
		checkHTTPGolden(t, handler, directory, wantStatus)
	})
}

// checkHTTPGolden sends the checked-in HTTP request through the production
// handler and compares its complete response with the reviewed golden file.
func checkHTTPGolden(t *testing.T, handler http.Handler, directory string, wantStatus int) {
	t.Helper()

	// Verify fixture signatures and time claims at their fixed valid instant.
	// Business-clock tests may advance timex independently of token expiry.
	previous := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time { return time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC) }
	defer func() { jwt.TimeFunc = previous }()

	input, err := os.ReadFile(filepath.Join(directory, "request.http"))
	if err != nil {
		t.Fatalf("read request: %v", err)
	}
	request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(input)))
	if err != nil {
		t.Fatalf("parse request: %v", err)
	}
	defer request.Body.Close()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	response := recorder.Result()
	defer response.Body.Close()

	if response.StatusCode != wantStatus {
		t.Errorf("HTTP status=%d, want %d", response.StatusCode, wantStatus)
	}

	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		t.Fatalf("dump response: %v", err)
	}
	got := fmt.Sprintf(">>> %s %s\n%s", request.Method, request.RequestURI, dump)
	got = strings.ReplaceAll(got, "\r\n", "\n")

	goldenPath := filepath.Join(directory, "golden.http")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Errorf("HTTP response differs from %s\n--- got ---\n%s\n--- want ---\n%s", goldenPath, got, want)
	}
}
