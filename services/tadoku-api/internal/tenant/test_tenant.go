package tenant

import (
	"fmt"
	"regexp"
	"strings"
)

var testIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]*-[0-9a-f]{8}$`)

// Obtain lifecycle keys through ParseTestTenant; the zero value cannot target provider state.
type TestKey struct {
	key Key
}

func ParseTestTenant(raw string) (TestKey, error) {
	key, err := Parse(raw)
	if err != nil {
		return TestKey{}, err
	}
	_, id, _ := strings.Cut(raw, "/")
	if key == Production() || !testIDPattern.MatchString(id) {
		return TestKey{}, fmt.Errorf("test tenant id must be a route ending in a hyphen and eight lowercase hexadecimal digits")
	}
	return TestKey{key: key}, nil
}

func (key TestKey) Key() Key {
	return key.key
}

func (key TestKey) String() string {
	return key.key.String()
}
