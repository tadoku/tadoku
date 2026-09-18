# Tadoku API migration log

- [ ] Remove all references to the “native” API; use “Tadoku API” instead. The distinction does not make sense long term.
- [ ] Design service authentication for queue workers when they need API access.
- [ ] Remove the v1 OpenAPI runtime/types dependency after the legacy services are retired.
- [ ] After the Tadoku API migration is complete, split all announcement, post, and page routes into separate admin and frontend routes.
- [ ] Move the announcements timestamp columns to `timestamptz` in a standalone migration.
- [ ] Add a covering announcements index on (`namespace`, `created_at desc`, `id desc`) where `deleted_at is null`.
- [ ] Reconsider the announcements primary key as (`namespace`, `id`) so IDs can be scoped to their namespace.
- [ ] After every endpoint owned by a legacy service has migrated, observe the replacement for one to seven days, then delete the legacy service together with its parity subtests and Bazel dependencies.
- [ ] After the Tadoku API migration is complete, make a final documentation pass and delete all references to this migration.
