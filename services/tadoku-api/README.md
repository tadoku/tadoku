# Tadoku API

Tadoku API is Tadoku's only backend. It serves every public HTTP operation,
which the gateway exposes under `/api`, owns the PostgreSQL schema and runs the
leaderboard outbox worker. Its canonical contract is
[`spec/openapi.yaml`](spec/openapi.yaml).

## Commands

Run these from the repository root. Always use `bazel`, never `go`. Tests need
disposable PostgreSQL and Valkey instances; see Testing below.

```sh
bazel build //services/tadoku-api/...
bazel test //services/tadoku-api/... --test_output=errors
bazel test //services/tadoku-api/e2e:e2e_test --test_output=errors  # one target
bazel run //:gazelle            # after adding Go files or changing imports
./scripts/generate-sqlc.sh      # after changing a SQL query
./scripts/generate-openapi.sh   # after changing spec/openapi.yaml
```

## Documentation

- [Overview](../../docs/docs/tadoku-api/index.md): API domains, code ownership
  and where the contract and persistence live.
- [Conventions](../../docs/docs/tadoku-api/conventions.md): layering, feature
  packages, write workflows, authorization, errors and business time.
- [Contract and OpenAPI](../../docs/docs/tadoku-api/contract.md): the canonical
  spec, compatibility rules and code generation.
- [Database and migrations](../../docs/docs/tadoku-api/database.md): standalone
  migrations, SQL style, sqlc and transactions.
- [Testing](../../docs/docs/tadoku-api/testing.md): test principles, test
  infrastructure and repository tests.
- [HTTP end-to-end tests](../../docs/docs/tadoku-api/http-e2e.md): golden
  cases, fixture tokens, relationships and the Kratos fixture.
- [User journeys](../../docs/docs/tadoku-api/user-journeys.md): chained
  scenarios for the important user flows.
- [Import boundaries](../../docs/docs/tadoku-api/import-boundaries.md): Bazel
  visibility and dependency checks.
- [Runtime configuration](../../docs/docs/tadoku-api/configuration.md):
  environment variables, startup, provider clients and the leaderboard worker.
- [Authorization](../../docs/docs/architecture/authorization.md): JWT verification,
  the ban gate and access checks.
- [Migration log](MIGRATION_LOG.md): deferred cleanup.
