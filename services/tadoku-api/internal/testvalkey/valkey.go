// Package testvalkey validates the shared disposable Valkey used by tests.
package testvalkey

import (
	"errors"
	"net/url"
	"os"
)

// URL returns a tightly scoped loopback URL suitable for destructive test keys.
func URL() (string, error) {
	raw := os.Getenv("TADOKU_TEST_VALKEY_URL")
	u, err := url.Parse(raw)
	if err != nil || u == nil {
		return "", errors.New("TADOKU_TEST_VALKEY_URL must be a disposable loopback Valkey URL")
	}
	if u.Scheme != "redis" ||
		(u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1") ||
		u.Port() == "" ||
		u.User != nil ||
		u.Path != "" ||
		u.RawPath != "" ||
		u.Opaque != "" ||
		u.RawQuery != "" ||
		u.Fragment != "" {
		return "", errors.New("TADOKU_TEST_VALKEY_URL must use redis:// on loopback with an explicit port and no credentials, path, query or fragment")
	}
	return raw, nil
}
