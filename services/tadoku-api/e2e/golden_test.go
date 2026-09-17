package e2e_test

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
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

	got, err := formatHTTPGolden(request, response)
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

// formatHTTPGolden renders the request line, status, sorted headers, and body.
// Connection is omitted so httptest's missing Content-Length cannot inject
// Connection: close. Content-Length and Transfer-Encoding are never added.
func formatHTTPGolden(request *http.Request, response *http.Response) (string, error) {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, ">>> %s %s\n", request.Method, request.RequestURI)

	text := response.Status
	if text == "" {
		text = http.StatusText(response.StatusCode)
		if text == "" {
			text = fmt.Sprintf("status code %d", response.StatusCode)
		}
	} else {
		text = strings.TrimPrefix(text, fmt.Sprintf("%d ", response.StatusCode))
	}
	fmt.Fprintf(&buf, "HTTP/%d.%d %03d %s\n", response.ProtoMajor, response.ProtoMinor, response.StatusCode, text)

	headers := response.Header.Clone()
	headers.Del("Connection")
	if err := headers.Write(&buf); err != nil {
		return "", err
	}
	buf.WriteByte('\n')

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	buf.Write(body)

	return strings.ReplaceAll(buf.String(), "\r\n", "\n"), nil
}
