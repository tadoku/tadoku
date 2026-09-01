## Immersion-API

## Valkey leaderboard cache

Valkey is an optional acceleration layer for unfiltered contest, yearly, and
global leaderboards. Postgres remains the source of truth.

- If Valkey is unavailable at startup, the API starts in degraded mode and the
  client reconnects on later operations.
- Unfiltered leaderboard reads fall back to the corresponding paginated
  Postgres query when a Valkey fetch or rebuild fails. Filtered leaderboard
  reads already query Postgres directly.
- Score-changing writes commit to Postgres and enqueue leaderboard outbox
  events. Failed cache updates remain pending and are retried after Valkey
  recovers.
- `/livez` checks process liveness and `/readyz` checks Postgres. Valkey health
  intentionally does not remove an otherwise usable API instance from service.

The Valkey URL must contain exactly one standalone address. Cache latency is
bounded by these settings:

| Environment variable | Default | Purpose |
| --- | --- | --- |
| `API_VALKEY_TIMEOUT` | `1s` | Maximum duration of a Valkey connection attempt or leaderboard operation. |

To verify degraded behavior in development, stop the `valkey-immersion`
resource, confirm the API remains ready and leaderboard requests return
Postgres-backed responses, create or update a score, then restore Valkey.
Without restarting `immersion-api`, confirm the outbox event is processed.

## Code generation

```sh
# Generate new api client
bazel run //services/immersion-api:api_gen

# Generate new version of OpenAPI client
bazel run //services/immersion-api/http/rest/openapi:codegen_gen
```
