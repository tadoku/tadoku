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

var fixtureInstant = time.Date(2026, 9, 12, 12, 0, 0, 0, time.UTC)

func APITestName(operation string, status int, description ...string) string {
	return fmt.Sprintf("%s/%d_%s", operation, status, strings.Join(description, "_"))
}

type implementation struct {
	name    string
	handler http.Handler
	skip    string
}

func atFixtureInstant(fn func()) {
	previous := jwt.TimeFunc
	jwt.TimeFunc = func() time.Time { return fixtureInstant }
	defer func() { jwt.TimeFunc = previous }()
	timex.TheWorld(fixtureInstant, fn)
}

func runCase(t *testing.T, s *suite, name string, want int, implementations ...implementation) {
	t.Helper()
	dir := filepath.Join("testdata", name)
	for _, impl := range implementations {
		t.Run(impl.name, func(t *testing.T) {
			if impl.skip != "" {
				t.Skip(impl.skip)
			}
			s.reset(t, dir)
			atFixtureInstant(func() { checkHTTPGolden(t, impl.handler, dir, want) })
			if s.proxied.Load() != 0 {
				t.Error("handler contacted an upstream")
			}
		})
	}
}

// checkHTTPGolden sends the checked-in HTTP request through the handler and
// compares its complete response with the reviewed golden file.
func checkHTTPGolden(t *testing.T, handler http.Handler, directory string, wantStatus int) {
	t.Helper()

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

	got, err := formatHTTPGolden(request, recorder)
	if err != nil {
		t.Fatalf("format response: %v", err)
	}

	goldenPath := filepath.Join(directory, "golden.http")
	want, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden: %v", err)
	}
	if got != string(want) {
		t.Errorf("HTTP response differs from %s\n--- got ---\n%s\n--- want ---\n%s", goldenPath, got, want)
	}
}

// formatHTTPGolden completes an unknown response length from the buffered body
// before serializing the response and normalizing HTTP line endings.
func formatHTTPGolden(request *http.Request, recorder *httptest.ResponseRecorder) (string, error) {
	response := recorder.Result()
	if response.ContentLength == -1 {
		response.ContentLength = int64(recorder.Body.Len())
	}

	dump, err := httputil.DumpResponse(response, true)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(">>> %s %s\n%s", request.Method, request.RequestURI, strings.ReplaceAll(string(dump), "\r\n", "\n")), nil
}

func TestFormatHTTPGolden(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		headers   http.Header
		body      string
		want      string
		wantError bool
	}{
		{
			name:    "body and repeated headers",
			status:  http.StatusOK,
			headers: http.Header{"Set-Cookie": {"a=1", "b=2"}},
			body:    "ok",
			want:    "HTTP/1.1 200 OK\nContent-Length: 2\nContent-Type: text/plain; charset=utf-8\nSet-Cookie: a=1\nSet-Cookie: b=2\n\nok",
		},
		{
			name:   "empty error",
			status: http.StatusBadRequest,
			want:   "HTTP/1.1 400 Bad Request\nContent-Length: 0\n\n",
		},
		{
			name:   "no content",
			status: http.StatusNoContent,
			want:   "HTTP/1.1 204 No Content\n\n",
		},
		{
			name:   "not modified",
			status: http.StatusNotModified,
			want:   "HTTP/1.1 304 Not Modified\n\n",
		},
		{
			name:    "explicit keep alive",
			status:  http.StatusOK,
			headers: http.Header{"Connection": {"keep-alive"}},
			body:    "ok",
			want:    "HTTP/1.1 200 OK\nContent-Length: 2\nConnection: keep-alive\nContent-Type: text/plain; charset=utf-8\n\nok",
		},
		{
			name:    "explicit close",
			status:  http.StatusOK,
			headers: http.Header{"Connection": {"close"}},
			body:    "ok",
			want:    "HTTP/1.1 200 OK\nContent-Length: 2\nConnection: close\nContent-Type: text/plain; charset=utf-8\n\nok",
		},
		{
			name:    "declared length",
			status:  http.StatusOK,
			headers: http.Header{"Content-Length": {"2"}},
			body:    "ok",
			want:    "HTTP/1.1 200 OK\nContent-Length: 2\nContent-Type: text/plain; charset=utf-8\n\nok",
		},
		{
			name:      "incorrect declared length",
			status:    http.StatusOK,
			headers:   http.Header{"Content-Length": {"5"}},
			body:      "ok",
			wantError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			recorder := httptest.NewRecorder()
			for name, values := range test.headers {
				for _, value := range values {
					recorder.Header().Add(name, value)
				}
			}
			if test.body != "" {
				recorder.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}
			recorder.WriteHeader(test.status)
			if test.body != "" {
				recorder.WriteString(test.body)
			}
			t.Cleanup(func() { recorder.Result().Body.Close() })

			got, err := formatHTTPGolden(request, recorder)
			if test.wantError {
				if err == nil {
					t.Fatalf("format response succeeded with an incorrect declared length: %q", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("format response: %v", err)
			}

			want := ">>> GET /\n" + test.want
			if got != want {
				t.Errorf("HTTP golden=%q, want %q", got, want)
			}
		})
	}
}
