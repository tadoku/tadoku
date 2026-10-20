package e2e_test

import (
	"github.com/tadoku/tadoku/services/tadoku-api/features/identity"
	"github.com/tadoku/tadoku/services/tadoku-api/internal/featureflags"
	valkeygo "github.com/valkey-io/valkey-go"
)

// routerOption replaces a dependency endpoint the fixtures normally supply.
type routerOption func(*routerEndpoints)

type routerEndpoints struct {
	jwks             string
	production       *bool
	ketoWrite        string
	identityValkey   valkeygo.Client
	identityObserver featureflags.Observer
	legacySessions   *identity.LegacySessions
	legacyReader     identity.LegacyReader
	flags            featureflags.BooleanProvider
}

func withIdentityValkey(client valkeygo.Client) routerOption {
	return func(endpoints *routerEndpoints) { endpoints.identityValkey = client }
}

func withKetoWriteURL(url string) routerOption {
	return func(endpoints *routerEndpoints) { endpoints.ketoWrite = url }
}

func withLegacyIdentity(sessions *identity.LegacySessions, reader identity.LegacyReader) routerOption {
	return func(endpoints *routerEndpoints) {
		endpoints.legacySessions = sessions
		endpoints.legacyReader = reader
	}
}

func withFlagProvider(provider featureflags.BooleanProvider) routerOption {
	return func(endpoints *routerEndpoints) { endpoints.flags = provider }
}
