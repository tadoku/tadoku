# Tadoku API migration log

- [ ] Decide how to replace the remaining “native” values: the proxy request metric `mode` label and the leaderboard cache marker.
- [ ] Design service authentication for queue workers when they need API access.
- [ ] After the Tadoku API migration is complete, split all announcement, post, and page routes into separate admin and frontend routes.
- [ ] Move the announcements timestamp columns to `timestamptz` in a standalone migration.
- [ ] Reconsider the announcements primary key as (`namespace`, `id`) so IDs can be scoped to their namespace.
- [ ] After the Tadoku API migration is complete, make a final documentation pass and delete all references to this migration.
- [x] Add the job step kind to the journey runner together with the first migrated worker. The outbox journey starts the production `Run` loop, waits for its ready signal and a later poll after an API write, then stops and joins it during cleanup.
- [ ] Maintain the Keto-backed contest-create permission through a background reconciliation job.
- [x] Restructure the Tadoku API documentation for progressive disclosure for agents: a short entry point with focused documents behind it instead of one long README.
- [ ] Rename the standard test identity from `Reader One` to `User One` across Kratos seeds, tests and HTTP goldens; regenerate affected signed JWT fixtures and update the public JWKS together.
- [ ] Rename `moderation_audit_log` to an audit-owned table in a standalone migration.
- [ ] Replace `github.com/google/uuid` with the standard-library UUID API when it is available in the adopted Go toolchain.
- [ ] Introduce a generic paginated request type and convert every existing paginated operation to use it.
- [ ] Support two-step audit recording for external actions: persist the start before calling the external system, then record the correlated completion and outcome. External changes cannot share an atomic PostgreSQL transaction with the audit write.
- [ ] Decide where shared repository row mappers belong.
- [ ] Introduce generic conversion helpers for repeated slice and type mappings after common conversion patterns stabilize across migrated features.
- [ ] Review and standardize the role and identity checking patterns used by application operations.
- [ ] Configure `wsl_v5` as a required CI check for handwritten Tadoku API Go code, excluding generated files; enable `after-block` and `after-decl` checks and provide a local auto-fix command.

- [ ] Review the log tag array-text decoding contract before replacing it with native array decoding; escaped quotes and backslashes currently affect returned tags.

- [ ] Add a domain-errors lint that not-found sentinels use `errx` (plain `errors.New` maps to Unknown→500).
- [ ] Ban `Normalize*` functions returning Internal errors (lint).
- [ ] Depolicy/sqlc lint: contests SQL must not declare catalog-only `from languages` without a contest join/filter.
- [ ] CI deny `ory/kratos-client-go` imports outside `features/profile`.
- [ ] Deny `valkey-go` imports outside Store code.
- [ ] Ban private unit→activity maps outside `domain/activities` (lint).
- [x] Share one profile Kratos traits decoder; remove the dead untagged `Email` field in `FindProfile`.
- [ ] Forbid raw banned/admins Keto triples outside `internal/permissions` (lint).
- [ ] Standardize caller UUID / self-or-admin helpers in app; ban ad-hoc Subject UUID parse outside allowlisted guest paths.
- [ ] Decide whether pages/posts should share content-revision primitives or stay intentional twins.
- [x] Bind the ban-middleware path carve-out to the mux/generated `AuthzRoleGet` pattern, not the magic `"/authz/current-user/role"` string.
