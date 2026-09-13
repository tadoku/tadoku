package http

import (
	"io"
	stdhttp "net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"
)

func TestFacadePreservesLegacyResponses(t *testing.T) {
	for _, status := range []int{200, 204, 302, 400, 401, 403, 404, 409, 422, 429, 500, 503} {
		t.Run(stdhttp.StatusText(status), func(t *testing.T) {
			upstream := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
				w.Header().Set("Content-Type", "application/problem+json")
				w.Header().Set("Location", "/next?return=%2Fcontests")
				w.Header().Set("Retry-After", "30")
				w.Header().Set("Cache-Control", "private, no-store")
				w.Header().Add("Set-Cookie", "first=one; HttpOnly; Secure")
				w.Header().Add("Set-Cookie", "second=two; SameSite=Lax")
				w.WriteHeader(status)
				if status != stdhttp.StatusNoContent {
					_, _ = io.WriteString(w, `{"detail":"legacy response"}`)
				}
			}))
			t.Cleanup(upstream.Close)
			facade := httptest.NewServer(newTestHandler(t, upstream.URL, 5*time.Second))
			t.Cleanup(facade.Close)
			client := &stdhttp.Client{
				Timeout:       5 * time.Second,
				CheckRedirect: func(*stdhttp.Request, []*stdhttp.Request) error { return stdhttp.ErrUseLastResponse },
			}
			direct, err := client.Get(upstream.URL + "/response")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer direct.Body.Close()
			proxied, err := client.Get(facade.URL + "/content/response")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			defer proxied.Body.Close()
			if proxied.StatusCode != direct.StatusCode {
				t.Errorf("got %v, want %v", proxied.StatusCode, direct.StatusCode)
			}
			for _, header := range []string{"Content-Type", "Location", "Retry-After", "Cache-Control", "Set-Cookie"} {
				if !slices.Equal(direct.Header.Values(header), proxied.Header.Values(header)) {
					t.Errorf("got %v, want %v", proxied.Header.Values(header), direct.Header.Values(header))
				}
			}
			directBody, err := io.ReadAll(direct.Body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			proxiedBody, err := io.ReadAll(proxied.Body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !slices.Equal(directBody, proxiedBody) {
				t.Errorf("got %v, want %v", proxiedBody, directBody)
			}
		})
	}
}

func TestFacadeStreamsResponseBeforeUpstreamFinishes(t *testing.T) {
	finish := make(chan struct{})
	upstream := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = io.WriteString(w, "data: first\n\n")
		_ = stdhttp.NewResponseController(w).Flush()
		select {
		case <-finish:
			_, _ = io.WriteString(w, "data: last\n\n")
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(upstream.Close)
	facade := httptest.NewServer(newTestHandler(t, upstream.URL, 5*time.Second))
	t.Cleanup(facade.Close)
	response, err := (&stdhttp.Client{Timeout: 5 * time.Second}).Get(facade.URL + "/immersion/events")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer response.Body.Close()
	first := make([]byte, len("data: first\n\n"))
	_, err = io.ReadFull(response.Body, first)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(first) != "data: first\n\n" {
		t.Errorf("got %v, want %v", string(first), "data: first\n\n")
	}
	close(finish)
	last, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(last) != "data: last\n\n" {
		t.Errorf("got %v, want %v", string(last), "data: last\n\n")
	}
}

func BenchmarkFacade(b *testing.B) {
	upstream := httptest.NewServer(stdhttp.HandlerFunc(func(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
		_, _ = io.WriteString(w, `{"id":"example","title":"A small API response"}`)
	}))
	b.Cleanup(upstream.Close)
	facade := httptest.NewServer(newTestHandler(b, upstream.URL, 5*time.Second))
	b.Cleanup(facade.Close)
	client := &stdhttp.Client{Timeout: 5 * time.Second}
	for _, route := range []struct{ name, url string }{
		{"direct", upstream.URL + "/pages/example"},
		{"proxy", facade.URL + "/content/pages/example"},
	} {
		b.Run(route.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				response, err := client.Get(route.url)
				if err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
				_, err = io.Copy(io.Discard, response.Body)
				_ = response.Body.Close()
				if err != nil {
					b.Fatalf("unexpected error: %v", err)
				}
				if response.StatusCode != stdhttp.StatusOK {
					b.Fatalf("got %v, want %v", response.StatusCode, stdhttp.StatusOK)
				}
			}
		})
	}
}
