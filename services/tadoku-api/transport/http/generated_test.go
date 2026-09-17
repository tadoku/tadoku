package http

import (
	"context"
	"encoding/json"
	"errors"
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/oapi-codegen/nullable"
	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func TestFindAnnouncementByIDInvalidUUID(t *testing.T) {
	s := &server{}
	response, err := s.ContentAnnouncementFindByID(context.Background(), openapi.ContentAnnouncementFindByIDRequestObject{
		Namespace: "tadoku",
		Id:        "not-a-uuid",
	})
	if !errors.Is(err, errInvalidUUID) {
		t.Fatalf("error = %v, want invalid UUID", err)
	}
	if response != nil {
		t.Errorf("response = %v, want nil", response)
	}
}

func TestDeleteAnnouncementInvalidUUID(t *testing.T) {
	s := &server{}
	response, err := s.ContentAnnouncementDelete(context.Background(), openapi.ContentAnnouncementDeleteRequestObject{
		Namespace: "tadoku",
		Id:        "not-a-uuid",
	})
	if !errors.Is(err, errInvalidUUID) {
		t.Fatalf("error = %v, want invalid UUID", err)
	}
	if response != nil {
		t.Errorf("response = %v, want nil", response)
	}
}

func TestUpdateAnnouncementInvalidUUID(t *testing.T) {
	s := &server{}
	response, err := s.ContentAnnouncementUpdate(context.Background(), openapi.ContentAnnouncementUpdateRequestObject{
		Namespace: "tadoku",
		Id:        "not-a-uuid",
		Body:      &openapi.ContentAnnouncementUpdateJSONRequestBody{},
	})
	if !errors.Is(err, errInvalidUUID) {
		t.Fatalf("error = %v, want invalid UUID", err)
	}
	if response != nil {
		t.Errorf("response = %v, want nil", response)
	}
}

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

func TestGeneratedNullableStringJSON(t *testing.T) {
	tests := []struct {
		name string
		href nullable.Nullable[string]
		want string
	}{
		{name: "null", href: nullable.NewNullNullable[string](), want: "null"},
		{name: "value", href: nullable.NewNullableWithValue("https://tadoku.app/announcement"), want: `"https://tadoku.app/announcement"`},
		{name: "empty string", href: nullable.NewNullableWithValue(""), want: `""`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(openapi.ContentAnnouncement{Href: test.href})
			if err != nil {
				t.Fatal(err)
			}

			var fields map[string]json.RawMessage
			if err := json.Unmarshal(body, &fields); err != nil {
				t.Fatal(err)
			}
			if got := string(fields["href"]); got != test.want {
				t.Errorf("href = %s, want %s", got, test.want)
			}
		})
	}
}
