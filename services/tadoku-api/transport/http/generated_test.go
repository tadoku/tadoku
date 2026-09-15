package http

import (
	stdhttp "net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONCharsetCompatibility(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		status      int
		want        string
	}{
		{
			name:        "generated JSON response",
			contentType: "application/json",
			status:      stdhttp.StatusOK,
			want:        "application/json; charset=UTF-8",
		},
		{
			name:        "non-JSON response",
			contentType: "text/plain; charset=utf-8",
			status:      stdhttp.StatusOK,
			want:        "text/plain; charset=utf-8",
		},
		{
			name:   "empty error response",
			status: stdhttp.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			handler := withJSONCharsetCompatibility(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
				if test.contentType != "" {
					w.Header().Set("Content-Type", test.contentType)
				}
				w.WriteHeader(test.status)
			}))

			handler.ServeHTTP(response, httptest.NewRequest(stdhttp.MethodGet, "/", nil))

			if got := response.Header().Get("Content-Type"); got != test.want {
				t.Errorf("Content-Type = %q, want %q", got, test.want)
			}
		})
	}
}

func TestJSONCharsetResponseWriterUnwraps(t *testing.T) {
	response := httptest.NewRecorder()
	w := &jsonCharsetResponseWriter{ResponseWriter: response}

	if got := w.Unwrap(); got != response {
		t.Errorf("Unwrap() = %T, want original response writer", got)
	}
}
