package e2e_test

import valkeygo "github.com/valkey-io/valkey-go"

// routerOption replaces a dependency endpoint the fixtures normally supply.
type routerOption func(*routerEndpoints)

type routerEndpoints struct {
	jwks           string
	production     *bool
	ketoWrite      string
	identityValkey valkeygo.Client
}

func withIdentityValkey(client valkeygo.Client) routerOption {
	return func(endpoints *routerEndpoints) { endpoints.identityValkey = client }
}

func withKetoWriteURL(url string) routerOption {
	return func(endpoints *routerEndpoints) { endpoints.ketoWrite = url }
}
