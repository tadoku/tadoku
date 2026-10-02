package tenant

import (
	"context"
	"fmt"
	"regexp"
)

var keyPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,55}/[a-z0-9][a-z0-9-]{0,55}$`)

// Obtain keys through Parse or Production; the zero value is not a tenant.
type Key struct {
	value string
}

func Parse(raw string) (Key, error) {
	if !keyPattern.MatchString(raw) {
		return Key{}, fmt.Errorf("tenant key must contain two lowercase name/id segments of at most 56 bytes each")
	}
	return Key{value: raw}, nil
}

func Production() Key {
	return Key{value: "tadoku/prod"}
}

func (key Key) String() string {
	return key.value
}

type contextKey struct{}

func WithKey(ctx context.Context, key Key) context.Context {
	return context.WithValue(ctx, contextKey{}, key)
}

func FromContext(ctx context.Context) (Key, bool) {
	key, ok := ctx.Value(contextKey{}).(Key)
	return key, ok && key.value != ""
}

type Deployment struct {
	key Key
}

func ParseDeployment(raw string) (Deployment, error) {
	if raw == "" {
		return Deployment{}, nil
	}

	key, err := Parse(raw)
	if err != nil {
		return Deployment{}, err
	}
	if key == Production() {
		return Deployment{}, fmt.Errorf("production tenant cannot scope a branch deployment")
	}
	return Deployment{key: key}, nil
}

func (deployment Deployment) Serves(key Key) bool {
	return key.value != "" && (deployment.key.value == "" || deployment.key == key)
}

func (deployment Deployment) Key() (Key, bool) {
	return deployment.key, deployment.key.value != ""
}
