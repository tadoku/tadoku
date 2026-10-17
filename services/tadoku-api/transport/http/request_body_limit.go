package http

import (
	stdhttp "net/http"
	"strings"
)

const (
	maxRequestBodyBytes         = 64 << 10
	maxDocumentRequestBodyBytes = 1 << 20
)

func withRequestBodyLimit(next stdhttp.Handler) stdhttp.Handler {
	return stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		r.Body = stdhttp.MaxBytesReader(w, r.Body, requestBodyLimit(r.Pattern))

		next.ServeHTTP(w, r)
	})
}

func requestBodyLimit(pattern string) int64 {
	_, path, _ := strings.Cut(pattern, " ")

	if strings.HasPrefix(path, "/content/pages/") || strings.HasPrefix(path, "/content/posts/") {
		return maxDocumentRequestBodyBytes
	}

	return maxRequestBodyBytes
}
