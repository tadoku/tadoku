package e2e_test

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
)

const (
	documentedRequestBodyLimit        = 64 << 10
	documentedContentRequestBodyLimit = 1 << 20
)

func TestRequestBodyLimit(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		body    string
		want    int
	}{
		{
			name:    "guest write over the default limit",
			fixture: APITestName("ImmersionLogCreate", http.StatusUnauthorized, "guest", "empty", "body"),
			body:    paddedJSONBody("description", documentedRequestBodyLimit+1),
			want:    http.StatusRequestEntityTooLarge,
		},
		{
			name:    "member write over the default limit",
			fixture: APITestName("ImmersionLogCreate", http.StatusBadRequest, "description", "too", "long"),
			body:    paddedJSONBody("description", documentedRequestBodyLimit+1),
			want:    http.StatusRequestEntityTooLarge,
		},
		{
			name:    "member write within the default limit reaches validation",
			fixture: APITestName("ImmersionLogCreate", http.StatusBadRequest, "description", "too", "long"),
			body:    paddedJSONBody("description", documentedRequestBodyLimit),
			want:    http.StatusBadRequest,
		},
		{
			name:    "callback write over the default limit",
			fixture: APITestName("AuthzProxyProxyAdminCheck", http.StatusOK, "admin"),
			body:    paddedJSONBody("subject", documentedRequestBodyLimit+1),
			want:    http.StatusRequestEntityTooLarge,
		},
		{
			name:    "CMS document within the content allowance",
			fixture: APITestName("CreatePage", http.StatusCreated, "admin"),
			body:    pageJSONBody(documentedContentRequestBodyLimit / 2),
			want:    http.StatusCreated,
		},
		{
			name:    "CMS document over the content allowance",
			fixture: APITestName("CreatePage", http.StatusCreated, "admin"),
			body:    pageJSONBody(documentedContentRequestBodyLimit + 1),
			want:    http.StatusRequestEntityTooLarge,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := filepath.Join("testdata", test.fixture)
			api.reset(t, directory)
			request := fixtureRequestWithBody(t, directory, test.body)
			response := httptest.NewRecorder()

			atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })

			if response.Code != test.want {
				t.Errorf("status=%d, want %d", response.Code, test.want)
			}
		})
	}
}

func TestRequestBodyLimitRejectsUndeclaredLength(t *testing.T) {
	fixture := APITestName("ImmersionLogCreate", http.StatusUnauthorized, "guest", "empty", "body")
	directory := filepath.Join("testdata", fixture)
	api.reset(t, directory)
	request := fixtureRequestWithBody(t, directory, paddedJSONBody("description", documentedRequestBodyLimit+1))
	request.ContentLength = -1
	request.Header.Del("Content-Length")
	response := httptest.NewRecorder()

	atFixtureInstant(func() { api.handler.ServeHTTP(response, request) })

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status=%d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
}

func fixtureRequestWithBody(t *testing.T, directory, body string) *http.Request {
	t.Helper()

	request := readHTTPRequest(t, directory)
	if err := request.Body.Close(); err != nil {
		t.Fatalf("close fixture body: %v", err)
	}
	request.Body = io.NopCloser(strings.NewReader(body))
	request.ContentLength = int64(len(body))
	request.Header.Set("Content-Length", fmt.Sprint(len(body)))

	return request
}

func paddedJSONBody(field string, size int) string {
	envelope := fmt.Sprintf(`{%q:""}`, field)

	return fmt.Sprintf(`{%q:%q}`, field, strings.Repeat("a", size-len(envelope)))
}

func pageJSONBody(size int) string {
	const template = `{"slug":"big-page","title":"Big page","html":%q}`
	envelope := fmt.Sprintf(template, "")

	return fmt.Sprintf(template, strings.Repeat("a", size-len(envelope)))
}
