package flipt

import (
	"context"
	"errors"
	"strings"

	"github.com/tadoku/tadoku/services/tadoku-api/internal/tenant"
)

type Target struct {
	Environment string
	Namespace   string
}

type Targets struct {
	production      Target
	testEnvironment string
}

func NewTargets(environment, namespace, testEnvironment string) (Targets, error) {
	if strings.TrimSpace(environment) == "" || strings.TrimSpace(namespace) == "" {
		return Targets{}, errors.New("Flipt production environment and namespace are required")
	}
	if strings.TrimSpace(testEnvironment) == "" || testEnvironment == environment {
		return Targets{}, errors.New("Flipt test environment must be nonempty and differ from production")
	}
	return Targets{
		production:      Target{Environment: environment, Namespace: namespace},
		testEnvironment: testEnvironment,
	}, nil
}

func (targets Targets) Resolve(ctx context.Context) (Target, error) {
	key, ok := tenant.FromContext(ctx)
	if !ok || targets.production.Environment == "" || targets.testEnvironment == "" {
		return Target{}, errors.New("Flipt target requires configured environments and a tenant")
	}
	if key == tenant.Production() {
		return targets.production, nil
	}
	return Target{
		Environment: targets.testEnvironment,
		Namespace:   strings.ReplaceAll(key.String(), "/", "_"),
	}, nil
}
