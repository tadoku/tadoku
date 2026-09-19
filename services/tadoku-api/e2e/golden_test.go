package e2e_test

import (
	"bufio"
	"bytes"
	"flag"
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
var updateGoldens = flag.Bool("update-goldens", false, "rewrite golden.http files from actual responses; review the diff before committing")

const goldenSourceRootEnv = "TADOKU_GOLDEN_SOURCE_ROOT"
const goldenRecorder = "tadoku-api"

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
			record := *updateGoldens && impl.name == goldenRecorder
			atFixtureInstant(func() { checkHTTPGolden(t, impl.handler, dir, want, record) })
			if s.proxied.Load() != 0 {
				t.Error("handler contacted an upstream")
			}
		})
	}
}

// checkHTTPGolden sends the checked-in HTTP request through the handler and
// compares or explicitly records its complete response.
func checkHTTPGolden(t *testing.T, handler http.Handler, directory string, wantStatus int, update bool) {
	t.Helper()
	checkHTTPResponseGolden(t, handler, readHTTPRequest(t, directory), directory, wantStatus, update)
}

// readHTTPRequest parses the directory's request.http with a fresh body.
func readHTTPRequest(t *testing.T, directory string) *http.Request {
	t.Helper()

	input, err := os.ReadFile(filepath.Join(directory, "request.http"))
	if err != nil {
		t.Fatalf("read request: %v", err)
	}
	request, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(input)))
	if err != nil {
		t.Fatalf("parse request: %v", err)
	}
	return request
}

// checkHTTPResponseGolden serves the request and compares or explicitly
// records the complete response against the directory's golden.http.
func checkHTTPResponseGolden(t *testing.T, handler http.Handler, request *http.Request, directory string, wantStatus int, update bool) {
	t.Helper()
	defer request.Body.Close()

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	response := recorder.Result()
	defer response.Body.Close()

	got, err := formatHTTPGolden(request, recorder)
	if err != nil {
		t.Fatalf("format response: %v", err)
	}

	goldenPath, err := goldenFilePath(directory, "golden.http", *updateGoldens)
	if err != nil {
		t.Fatal(err)
	}
	updated, err := reconcileHTTPGolden(goldenPath, got, response.StatusCode, wantStatus, update)
	if err != nil {
		t.Error(err)
		return
	}
	if updated {
		fmt.Fprintf(os.Stderr, "rewrote HTTP golden %s\n", goldenPath)
	}
}

func goldenFilePath(directory, name string, update bool) (string, error) {
	if !update {
		return filepath.Join(directory, name), nil
	}

	root := os.Getenv(goldenSourceRootEnv)
	if !filepath.IsAbs(root) {
		return "", fmt.Errorf("%s must be an absolute path to services/tadoku-api/e2e/testdata", goldenSourceRootEnv)
	}
	relative, err := filepath.Rel("testdata", directory)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("golden directory %q is outside testdata", directory)
	}

	return filepath.Join(root, relative, name), nil
}

func reconcileHTTPGolden(path, got string, gotStatus, wantStatus int, update bool) (bool, error) {
	if gotStatus != wantStatus {
		return false, fmt.Errorf("HTTP status=%d, want %d", gotStatus, wantStatus)
	}
	return reconcileGolden(path, got, update)
}

// reconcileGolden compares got with the checked-in golden, or rewrites an
// existing golden in update mode. It never creates a missing file.
func reconcileGolden(path, got string, update bool) (bool, error) {
	want, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read golden: %w", err)
	}
	if update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			return false, fmt.Errorf("write golden: %w", err)
		}
		return true, nil
	}
	if got != string(want) {
		return false, fmt.Errorf("golden %s differs\n--- got ---\n%s\n--- want ---\n%s", path, got, want)
	}

	return false, nil
}

// formatHTTPGolden completes an unknown response length from the buffered body
// before serializing the response and normalizing HTTP line endings.
func formatHTTPGolden(request *http.Request, recorder *httptest.ResponseRecorder) (string, error) {
	response := recorder.Result()
	if http.StatusText(response.StatusCode) == "" {
		response.Status = ""
	}
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
			name:   "unknown status text",
			status: 499,
			want:   "HTTP/1.1 499 status code 499\nContent-Length: 0\n\n",
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

func TestReconcileHTTPGolden(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "golden.http")
	if err := os.WriteFile(path, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("updates existing fixture", func(t *testing.T) {
		updated, err := reconcileHTTPGolden(path, "new", http.StatusOK, http.StatusOK, true)
		if err != nil {
			t.Fatal(err)
		}
		if !updated {
			t.Error("golden was not reported as updated")
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != "new" {
			t.Errorf("golden=%q, want %q", got, "new")
		}
	})

	t.Run("comparison never overwrites", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "golden.http")
		if err := os.WriteFile(path, []byte("new"), 0o644); err != nil {
			t.Fatal(err)
		}

		updated, err := reconcileHTTPGolden(path, "legacy", http.StatusOK, http.StatusOK, false)
		if err == nil {
			t.Fatal("mismatched golden comparison succeeded")
		}
		if updated {
			t.Error("comparison reported an update")
		}
		got, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(got) != "new" {
			t.Errorf("golden=%q, want %q", got, "new")
		}
	})

	t.Run("unexpected status never overwrites", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "golden.http")
		if err := os.WriteFile(path, []byte("new"), 0o644); err != nil {
			t.Fatal(err)
		}

		updated, err := reconcileHTTPGolden(path, "unexpected", http.StatusBadRequest, http.StatusOK, true)
		if err == nil {
			t.Fatal("unexpected status succeeded")
		}
		if updated {
			t.Error("unexpected status reported an update")
		}
		got, readErr := os.ReadFile(path)
		if readErr != nil {
			t.Fatal(readErr)
		}
		if string(got) != "new" {
			t.Errorf("golden=%q, want %q", got, "new")
		}
	})

	t.Run("missing fixture is not created", func(t *testing.T) {
		missing := filepath.Join(directory, "missing.http")
		updated, err := reconcileHTTPGolden(missing, "new", http.StatusOK, http.StatusOK, true)
		if err == nil {
			t.Fatal("missing golden update succeeded")
		}
		if updated {
			t.Error("missing golden reported an update")
		}
		if _, statErr := os.Stat(missing); !os.IsNotExist(statErr) {
			t.Fatalf("missing golden stat error=%v, want not exist", statErr)
		}
	})
}

func TestGoldenFilePathUsesSourceRootForUpdates(t *testing.T) {
	root := t.TempDir()
	t.Setenv(goldenSourceRootEnv, root)

	got, err := goldenFilePath(filepath.Join("testdata", "Operation", "200_case"), "golden.http", true)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "Operation", "200_case", "golden.http")
	if got != want {
		t.Errorf("golden path=%q, want %q", got, want)
	}
}
