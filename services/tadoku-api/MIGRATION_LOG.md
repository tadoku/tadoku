# Tadoku API migration log

- [ ] Remove all references to the “native” API; use “Tadoku API” instead. The distinction does not make sense long term.
- [ ] Design service authentication for queue workers when they need API access.
- [ ] Remove the v1 OpenAPI runtime/types dependency after the legacy services are retired.
- [ ] After the Tadoku API migration is complete, split all announcement, post, and page routes into separate admin and frontend routes.
- [ ] Move the announcements timestamp columns to `timestamptz` in a standalone migration.
- [ ] Add a covering announcements index on (`namespace`, `created_at desc`, `id desc`) where `deleted_at is null`.
- [ ] Reconsider the announcements primary key as (`namespace`, `id`) so IDs can be scoped to their namespace.
- [ ] After every endpoint owned by a legacy service has migrated, require operational acceptance and evidence that no active caller uses the legacy service, then delete it together with its parity subtests and Bazel dependencies.
- [ ] After the Tadoku API migration is complete, make a final documentation pass and delete all references to this migration.
- [ ] Add the job step kind to the journey runner together with the first migrated worker. Workers expose one synchronous pass as a method returning an error; production `Run` loops over it and tests never start the loop.
- [ ] Restructure the Tadoku API documentation for progressive disclosure for agents: a short entry point with focused documents behind it instead of one long README.
- [ ] Rename the standard test identity from `Reader One` to `User One` across Kratos seeds, tests and HTTP goldens; regenerate affected signed JWT fixtures and update the public JWKS together.
- [ ] Rename `moderation_audit_log` to an audit-owned table in a standalone migration after the legacy authorization service is retired.
- [ ] Replace `github.com/google/uuid` with the standard-library UUID API when it is available in the adopted Go toolchain.
- [ ] Support two-step audit recording for external actions: persist the start before calling the external system, then record the correlated completion and outcome. External changes cannot share an atomic PostgreSQL transaction with the audit write.
- [ ] Remove the temporary exact legacy proxy registration for `/immersion/contests/ongoing-registrations` when it becomes native.
