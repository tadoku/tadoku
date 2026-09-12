# Tadoku API: first native slice

`GET /content/announcements/{namespace}/active` is implemented natively, by
default. Other operations retain the existing proxy ownership, including HEAD
and OPTIONS on that path. A native failure never falls back to Content.
Externally the gateway adds `/api`; its routes and audiences are unchanged.

## Code ownership

```
transport/http -> app -> features/content -> generated/sqlc/content
                     -> features/access  -> existing Keto client
cmd/tadoku-api constructs and closes the shared pgx/v5 pool and HTTP resources
```

Content service chooses one `timex.Now()` cutoff and a limit of ten. Its concrete
repository only queries and maps rows. `postgres.Executor` makes the repository
usable inside a future app-owned transaction; this read does not open an
unnecessary transaction. There are no new feature/repository interfaces.

## Contract and compatibility

`spec/openapi.yaml` is the one canonical contract: 72 public operations plus nine
retained internal/callback operations. Source-prefixed operation IDs and component
names avoid collisions. Equivalent Content page/post templates share one canonical
path; their old parameter names are recorded for compatibility. This changes no
wire URLs. Original upstream server/path metadata preserves the inventory of
direct internal callers; the merge does not make those routes public.

Run `./scripts/generate-openapi.sh` for the single framework-free DTO package.
The generator is the existing oapi-codegen v1.12.4 pinned in `go.mod`, now runnable
through Bazel. Its Bazel binary has no Go module build-info header; the pin is in
`go.mod`, not the generated header. Run `./scripts/generate-sqlc.sh` for SQL output.
The native query uses sqlc v1.31.1/pgx-v5; legacy generators are unchanged.

Documentation builds filtered public views from the canonical contract. They
retain the four existing documentation sections/URLs, not four independent API
contract sources. Legacy specs are frozen build and contract-comparison inputs.
The contract tests compare all 81 operations and their component/security shapes.
Manually registered health and metrics endpoints are separately inventoried in
the contract's `x-tadoku-operational-surfaces` metadata, not rendered as product API.

The native endpoint preserves Content's existing middleware behavior:

| Caller/result | Response |
| --- | --- |
| Gateway-issued guest JWT, ordinary user, admin | 200 |
| Missing/malformed bearer header | 400, legacy JSON error |
| Invalid/expired JWT | 401, legacy JSON error |
| Service JWT without the `content-api` audience | 403, empty body |
| Banned user (including banned admin) | 403, empty body |
| Keto evaluation unavailable | Existing fail-open public-read policy |
| Valid user JWT missing `iat` | Legacy 500 JSON, without reproducing the panic |
| Database/read failure | 500, empty body; no fallback |

The current legacy verifier does **not** enforce an issuer and does not configure
background JWKS refresh. This slice preserves both behaviors; changing them is
separate security/provider work. Caller-supplied identity headers are not trusted.
Do not reuse the public-read fail-open rule for administrative operations.

## Runtime configuration

In addition to the existing four upstream URLs, startup now requires:

- `API_JWKS`, `API_KETO_READ_URL` (same endpoints used by Content).
- Individual `API_POSTGRES_HOST`, `PORT` (default 5432), `DATABASE`, `USER`,
  `PASSWORD`, `SSLMODE` fields. `API_POSTGRES_URL` remains rejected.
- `API_POSTGRES_MAX_CONNECTIONS` (default 4, validated range 1–32).

Startup pings PostgreSQL and fetches signing keys before opening listeners.
`/readyz` checks PostgreSQL; `/livez` remains independent of dependency health.
Native route metrics use bounded labels in
`tadoku_api_native_request_duration_seconds`; the existing proxy metrics remain.
Shutdown closes request/metrics listeners, the pool, JWKS resources and idle HTTP
connections. The dev deployment uses the existing disposable development DB role;
secret synchronization and reset scripts include Tadoku API.

**Production activation is not part of this change.** Provision a dedicated
runtime credential with connect/schema-usage/select-on-announcements grants, not
a migration/admin DSN. Confirm provider TLS/pooling settings and the coexistence
connection budget: two native replicas default to eight connections, in addition
to the still-running legacy pools. Update production secrets/manifests and complete
the master rollout gate under separate release authorization before deployment.
No schema migration or write-owner handoff is required for this read.

## Verification

Native integration tests require `TADOKU_TEST_POSTGRES_URL` pointing to an explicit
loopback port, database `postgres`, credentials `postgres:postgres` and exactly
`sslmode=disable`. Missing or unsafe configuration fails before any connection;
new tests never skip. Fixtures create random disposable databases and apply the
complete canonical migration history. Do not point them at shared dev or production.

```sh
bazel test //services/tadoku-api/... --test_output=errors
bazel test //services/tadoku-api/e2e:e2e_test --test_output=all \
  --test_arg=-test.run=^$ --test_arg=-test.bench=BenchmarkAnnouncementRead \
  --test_arg=-test.benchtime=1000x
```

E2E tests assemble production constructors and real JWT/Keto adapters against
controlled HTTP responses, plus real PostgreSQL. Only the dedicated test-only
`e2e/legacy` package imports Echo/legacy HTTP bindings. Bazel visibility restricts
native local dependencies. CI checks the transitive runtime Echo ban, subtree
Testify ban, OpenAPI generation, sqlc generation and the database suites.

**Remaining architecture gate:** Depolicy has not been added: the inspected
upstream revision has no license. Bazel visibility and graph checks are useful
but are not a replacement claim for all planned import rules. Adoption still
needs permission, explicit configuration validation, positive/negative fixtures
and proof that test-file imports are checked. Same-package responsibilities and
business signatures still require review.
