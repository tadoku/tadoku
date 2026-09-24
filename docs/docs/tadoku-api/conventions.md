---
title: Conventions
description: Layering, feature package, write workflow, caller authorization, dependency, error, validation, business-time and readability rules for Tadoku API code.
sidebar_position: 2
---

# Conventions

Read this when you add or change an application operation, a feature service or
repository, or a shared domain package.

These rules apply to every operation. The layers themselves are summarized in
[Code ownership](./index.md#code-ownership).

## Applications and features

- Group feature operations by the business data they own, not by the page or
  URL that exposes them. Aggregates belong with their source data; application
  operations compose those results with independently owned identity or catalog
  data.
- Application operations compose features. Feature services own business
  decisions; repositories only query and map rows.
- Compose independent features in the application layer. Features never import
  or call sibling features.

## Shared domain packages

- Put business values and pure rules used by multiple features in
  `services/tadoku-api/domain/<concept>`. Application operations and features
  may import these packages.
- Put fixed business values shared across features in a domain package instead
  of embedding them in a write operation.
- Shared domain packages must not import application, feature, transport,
  generated, infrastructure or `internal` packages. Keep services, repositories
  and provider APIs out of them.
- Keep types, errors and validation used by only one feature in that feature's
  `domain.go`. Sharing a concept does not require moving its feature
  operations.
- Technical support concerns, such as business-time control and request
  identity, stay under `internal/`.
- Changing a layer boundary follows [Import boundaries](./import-boundaries.md).

## Feature package layout

Group a feature's operations by responsibility, not one file per operation:

| File | Contains |
| --- | --- |
| `domain.go` | Feature domain types, errors and shared validation. Declare the feature's domain errors together in one `var` block and reuse them from the service and repository. |
| `<feature>_service.go` | The service struct, its constructor and the business operations. |
| `<feature>_service_test.go` | Service and validation tests, always database-free. |
| `<feature>_repository.go` | The repository struct, its constructor and the queries. |
| `<feature>_repository_test.go` | The feature's only database tests, which exercise the repository directly. |

Service orchestration is exercised through HTTP E2Es; see [Testing](./testing.md).

## Write workflows

- The application authorizes the caller, coordinates locks and transactions
  across features, and composes their results.
- A feature service validates or normalizes its own inputs and sequences its
  own repository calls, including related rows and outbox writes.
- Do not add feature service methods that only pass a repository call through
  so an application operation can assemble that feature's write.
- When several features take part in a write, the application passes shared
  domain values between them without interpreting scoring or persistence
  details. The service that owns a business decision creates its result; the
  service that owns the data translates that result into its stored rows.
- Transaction mechanics are in [Transactions](./database.md#transactions).

## Routes and caller authorization

[Authorization](../architecture/authorization.md) documents the shared JWT and ban
pipeline, the `permissions.Checker` methods and the error-to-status mapping.
The following rules decide where each check belongs.

- Register application routes through the router's `Handle` or `HandleFunc`
  methods during construction. Every such route inherits the shared request
  deadline, authentication and ban check.
- Application operations own full caller-access checks, such as
  authenticated-user and administrator requirements. They receive a named
  `*permissions.Checker` (`internal/permissions`) and call its checks
  explicitly. Do not enforce administrator access with HTTP middleware.
- The actor is the authenticated user performing an operation; audit events
  record it as `ActorID`. Read the actor's user ID with `identity.ActorID`,
  which reports no actor for a missing identity or a guest, or with
  `identity.RequireActorID`, which returns unauthorized in those cases. Do not
  parse the identity subject in application operations.
- When the owner of a resource or an administrator may act on it, call the
  application's `requireOwnerOrAdmin` with the owner's user ID after
  `RequireAuthenticated`. It passes when the actor is the owner and otherwise
  requires administrator access.
- Construction binds the checker to the request-scoped `app:tadoku#admins` Keto
  lookup. The checker derives its subject from the verified `internal/identity`
  context and does not cache results.
- Do not duplicate the shared ban gate in application operations or feature
  services. Code that invokes an application operation outside the application
  router must provide an equivalent baseline ban policy.
- A feature service may inspect authorization facts only to expand behavior
  inside an operation the application has already authorized.
- Feature services read and change facts about other users through the concrete
  services in `services/common/authz/roles`. Target facts are never caller
  authorization: the operation still uses the checker for its own access
  decision. Batch facts are read for each request, not cached.
- Public permission checks use a typed, construction-time allowlist keyed by
  namespace and relation; production supplies an empty list. After
  authentication, input validation and allowlist membership, the checker
  evaluates the requested object in Keto for the verified request subject and
  returns the provider decision. Provider failures return unavailable.
- A trusted callback route authenticates its own required bearer credential
  before request decoding and records a callback-authentication fact
  (`internal/callbackauth`) that its application operation must require. It
  never creates a user identity or enters the JWT and ban pipeline. A subject in
  its body is a target for a provider fact lookup, never the caller.

## Repositories and stores

- Name a type that reads or writes authoritative data in PostgreSQL a
  *Repository*. Name a type that accesses auxiliary, non-authoritative storage,
  such as Valkey caches, derived data, pub/sub or coordination state, a
  *Store*.
- Repositories are concrete structs in
  `features/<feature>/<feature>_repository.go`. They reach PostgreSQL through
  `postgres.Executor` (`services/tadoku-api/infra/postgres/`) and the feature's generated sqlc
  package, and convert between sqlc rows and feature domain types internally.
- The leaderboard feature issues its Valkey cache commands directly with the
  raw client from `services/tadoku-api/infra/valkey/`; no Store type exists yet.
- Keep each repository method to one SQL statement. A coherent join or CTE
  counts as one statement and is appropriate when the data needs one database
  snapshot. Compose independent repository reads and writes in the feature
  service or application layer, inside an application-owned transaction when
  the operation must commit atomically. `tools/ci/repopolicy` enforces this in
  CI: handwritten repository functions issue at most one database statement and
  never allocate IDs or control transactions.

## Dependencies

- Pass concrete application-owned collaborators through constructors.
- Introduce a narrow, consumer-owned interface only for a genuine provider or
  layer boundary with multiple implementations. Never add an interface, including
  a feature or repository interface, solely to inject mocks. Exercise concrete
  providers through their boundary tests and HTTP E2Es.
- Provider-backed feature services may consume the shared feature-flag
  evaluator and an application-owned provider client. Keep provider types and
  identity adaptation inside the owning feature package.
- Provider-backed read caches:
  - reuse the common provider client's cursor support, preserve provider order,
    read every cursor page and reject repeated continuation tokens;
  - make no provider request at construction. The first operation that needs
    the cache loads it with its request context; later operations reuse that
    snapshot for five minutes, and refreshes are serialized;
  - keep the last complete snapshot when a refresh fails. An initial failure
    returns unavailable because no complete snapshot exists yet.

## Errors

- Application errors use `internal/errx.Error`, which carries a
  transport-neutral `Kind`, a message and an optional cause.
- Create them with named constructors such as
  `errx.NewInvalidInputError(message)` or
  `errx.NewUnavailableError(message, cause)`; the latter accepts `nil` when
  there is no underlying cause.
- `errx.KindOf` reads the outermost typed error with `errors.As`. The HTTP
  boundary maps its kind to a status; unknown or unclassified errors return 500.
- Ordinary `%w` wrapping preserves the metadata, and `Unwrap` preserves causes
  for `errors.Is` and `errors.As`. Do not encode categories in error text or
  wrap category sentinels.
- Add call-site context only when it contributes useful diagnostics.

## Validation

- Return specific validation errors: one validation condition per branch, a
  direct call to `errx.NewInvalidInputError`, and a message that identifies the
  invalid field or rule. Combine conditions only when they intentionally share
  one message.
- Do not create per-validation sentinel errors or error types.
- Use `Validate` for checks that leave request data unchanged and `Normalize`
  for canonicalization of local service or persistence input. Do not use vague
  names such as `Prepare`, and do not make validation functions return a
  rewritten request.

## Request types

When a feature request type has fields that only the feature sets, such as the
caller ID from the verified identity or a year from business time, make those
fields unexported and add getters. The feature package writes them directly
(`req.userID = ...`); other packages read them through getters
(`req.UserID()`), so transport and application code cannot set them. Fields
that callers legitimately set, such as a contest ID or language codes, stay
exported.

## Business time

- Never call `time.Now()` for business time; use `internal/timex.Now()`. This
  applies to services, repositories and background workers.
- Tests control business time with `timex.TheWorld`; see
  [Time in tests](./http-e2e.md#time-in-tests).
- Real time that only drives expiry, deadlines or latency, such as token-cache
  expiry or request duration, reads `time.Now()` and stays independent of
  business time.
- A shared `services/common` package that needs controllable real time gives
  the struct an unexported `now func() time.Time` field that defaults to
  `time.Now`. Only tests in the same package override it.

## Readability

- Separate setup, execution, error handling and response mapping with
  whitespace. Within a function, put a blank line between coherent phases such
  as authorization, input extraction, persistence and response mapping.
- Put unrelated struct fields and composite-literal entries on separate lines.
- Split application and transport operations into files by functionality, and
  give operations descriptive names.
- Keep constructors and resource lifecycle code visibly separate from endpoint
  behavior.

## Changes and documentation

- Keep each change small and reviewable. Implement the requested operation
  without unrelated service identities, audience checks or authorization
  features. Change authentication and shared ban enforcement in separately
  reviewed changes.
- Document durable conventions for the whole application: architecture,
  compatibility and testing rules that apply to all operations. Do not document
  individual endpoint implementations, enumerate their coverage, or add
  endpoint-status sections to general documentation. Use generic examples.
- Keep deferred cleanup notes as checkboxes in
  `services/tadoku-api/MIGRATION_LOG.md`.
