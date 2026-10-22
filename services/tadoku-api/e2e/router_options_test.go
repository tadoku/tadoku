package e2e_test

import (
	"github.com/jackc/pgx/v5/pgxpool"
	valkeygo "github.com/valkey-io/valkey-go"
)

// routerOption replaces a dependency endpoint the fixtures normally supply.
type routerOption func(*routerEndpoints)

type routerEndpoints struct {
	jwks           string
	production     *bool
	permissions    *pgxpool.Pool
	identityValkey valkeygo.Client
}

func withIdentityValkey(client valkeygo.Client) routerOption {
	return func(endpoints *routerEndpoints) { endpoints.identityValkey = client }
}

// withPermissionsPool serves permission checks from pool instead of the router's database pool.
func withPermissionsPool(pool *pgxpool.Pool) routerOption {
	return func(endpoints *routerEndpoints) { endpoints.permissions = pool }
}
