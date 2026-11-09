package e2e_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

const gravatarTimeout = 200 * time.Millisecond

type gravatarAvatar struct {
	status      int
	contentType string
	body        []byte
	hang        bool
}

var gravatarAvatars = map[string]gravatarAvatar{
	"reader@example.test":    {status: http.StatusOK, contentType: "image/png", body: []byte("reader avatar")},
	"mixed@example.test":     {status: http.StatusOK, contentType: "image/jpeg", body: []byte("mixed avatar")},
	"deleted@example.test":   {status: http.StatusOK, contentType: "image/png", body: []byte("deleted avatar")},
	"error@example.test":     {status: http.StatusInternalServerError},
	"svg@example.test":       {status: http.StatusOK, contentType: "image/svg+xml", body: []byte("<svg/>")},
	"oversized@example.test": {status: http.StatusOK, contentType: "image/png", body: bytes.Repeat([]byte("a"), 256<<10+1)},
	"slow@example.test":      {hang: true},

	"refresh@example.test":        {status: http.StatusOK, contentType: "image/png", body: []byte("refreshed avatar")},
	"refresh-error@example.test":  {status: http.StatusInternalServerError},
	"refresh-slow@example.test":   {hang: true},
	"journey-avatar@example.test": {status: http.StatusOK, contentType: "image/webp", body: []byte("fresh avatar")},
	"journey-error@example.test":  {status: http.StatusInternalServerError},
}

// gravatarFixture serves gravatarAvatars by email hash, answers 404 for any other hash and 400 for a request that
// differs from the one the API must send.
type gravatarFixture struct {
	server  *httptest.Server
	avatars map[string]gravatarAvatar

	mu    sync.Mutex
	calls map[string]int
}

func newGravatarFixture() *gravatarFixture {
	f := &gravatarFixture{avatars: make(map[string]gravatarAvatar), calls: make(map[string]int)}
	for email, avatar := range gravatarAvatars {
		f.avatars[gravatarHash(email)] = avatar
	}
	f.server = httptest.NewServer(http.HandlerFunc(f.serve))
	return f
}

func (f *gravatarFixture) serve(w http.ResponseWriter, r *http.Request) {
	hash, ok := strings.CutPrefix(r.URL.Path, "/avatar/")
	if !ok ||
		r.Method != http.MethodGet ||
		r.URL.RawQuery != "s=128&d=404" ||
		r.Header.Get("Authorization") != "" ||
		r.Header.Get("Cookie") != "" ||
		r.Header.Get("X-Forwarded-For") != "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	f.mu.Lock()
	f.calls[hash]++
	f.mu.Unlock()

	avatar, ok := f.avatars[hash]
	switch {
	case !ok:
		w.WriteHeader(http.StatusNotFound)
	case avatar.hang:
		<-r.Context().Done()
	default:
		w.Header().Set("Content-Type", avatar.contentType)
		w.WriteHeader(avatar.status)
		_, _ = w.Write(avatar.body)
	}
}

func (f *gravatarFixture) callsFor(email string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[gravatarHash(email)]
}

func (f *gravatarFixture) close() { f.server.Close() }

func gravatarHash(email string) string {
	sum := sha256.Sum256([]byte(email))
	return hex.EncodeToString(sum[:])
}
