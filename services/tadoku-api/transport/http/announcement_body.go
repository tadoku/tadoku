package http

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
	"strings"

	"github.com/tadoku/tadoku/services/tadoku-api/generated/openapi"
)

func isAnnouncementWriteRequest(r *http.Request) bool {
	return r.Pattern == "POST /content/announcements/{namespace}"
}

// Match the legacy binder's empty-body and media-type behavior before generated
// JSON decoding. Authentication and ban enforcement have already run.
func withAnnouncementBodyCompatibility(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isAnnouncementWriteRequest(r) {
			if err := announcementBody(r); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func announcementBody(r *http.Request) error {
	if r.ContentLength == 0 {
		r.Body = io.NopCloser(strings.NewReader("{}"))
		return nil
	}

	contentType := r.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(contentType, "application/json"):
		return nil
	case strings.HasPrefix(contentType, "application/xml"), strings.HasPrefix(contentType, "text/xml"):
		// The legacy model uses a pointer for href; nullable's map representation
		// is JSON-specific, so shadow that field while decoding legacy XML.
		var body struct {
			openapi.ContentAnnouncement
			Href *string `json:"href"`
		}
		if err := xml.NewDecoder(r.Body).Decode(&body); err != nil {
			return err
		}
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		r.Body = io.NopCloser(bytes.NewReader(encoded))
		return nil
	case strings.HasPrefix(contentType, "application/x-www-form-urlencoded"), strings.HasPrefix(contentType, "multipart/form-data"):
		var err error
		if strings.HasPrefix(contentType, "multipart/form-data") {
			err = r.ParseMultipartForm(32 << 20)
		} else {
			err = r.ParseForm()
		}
		if err != nil {
			return err
		}
		// The generated legacy model has no form tags, so no fields bind.
		r.Body = io.NopCloser(strings.NewReader("{}"))
		return nil
	default:
		return http.ErrNotSupported
	}
}
