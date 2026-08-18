---
sidebar_position: 5
title: Scoring Rules
---

# Scoring rules

## Status

This document tracks the implementation of
[ADR 005: Scoring rules and metadata ownership](../../docs/adr/005-scoring-rules.md).

The work should be delivered as a series of small, deployable changes on `main`.
Published scores must remain snapshots: activating a new rule-set version must
not silently change scores for existing logs or contest submissions.

## Rule evaluation semantics

- Rules are evaluated by explicit ascending priority, with more specific rules
  ordered before broader rules.
- A non-stackable rule is a base rule. The first matching base rule supplies
  the score rate; other matching base rules are ignored.
- A stackable rule is a modifier. After selecting a base, every matching
  modifier multiplies the score. Modifiers do not match or score on their own
  when no base rule matches.
- Every populated matcher on a rule must match.
- Rules may match activity, stable unit key, language, and one normalized tag.
- Rules declare a `score_source`:
  - `amount` multiplies the submitted amount by the rule rate.
  - `duration_minutes` converts the submitted duration to minutes and multiplies
    it by the rule rate.
- Logs containing both amount and duration use `amount` as their score source.
  Duration remains additional tracking metadata.
- An unmatched rule produces a score of zero. It does not reject the log.
- Published rule sets are immutable and versioned.

For example:

```text
amount: 200 pages * 1.0 = 200
duration_minutes: 30 minutes * 0.4 = 12
dense listening: 30 minutes * 0.4 base * 1.5 modifier = 18
```

## Rule-set behavior

The platform has an active, versioned rule set. A contest can use one of two
modes:

- `override`: evaluate the contest rules first, then fall back to the platform
  rule-set version pinned by the contest.
- `replace`: evaluate only the contest rules. An unmatched submission scores
  zero for that contest.

Pinning the platform fallback prevents a platform rule change from altering a
contest's scoring behavior while the contest is running.

## Current implementation

Activity time tracking is already on `main`, but its V2 UI remains behind the
existing feature gate. The original non-admin logging form still submits the
legacy amount and UUID unit pair.

The backend currently works as follows:

- Activities are code-owned in
  `services/immersion-api/domain/activities.go`.
- Units are rows in `data.log_units`. They use generated UUIDs and contain
  scoring modifiers and optional language overrides.
- `LogCreate` and `LogUpdate` call `resolveLogTracking`, which loads the
  selected unit and calls `ComputeInterimLogScore`.
- `ComputeInterimLogScore` is in
  `services/immersion-api/domain/logtracking.go`.
- Amount-plus-duration logs are scored from amount and unit modifier. Duration
  is metadata.
- Duration-only logs use hard-coded per-minute rates in the interim scorer.
- New scores are written to `computed_score`.
- Historical amount logs may still rely on the generated `score` column.
- Read and aggregation queries use `coalesce(computed_score, score)`.
- `contest_logs` contains a separate score snapshot, but currently receives the
  same tracking and score snapshot as the base log.
- Updating a log rewrites snapshots for ongoing contests only. Completed
  contest snapshots are not updated.
- Leaderboard refreshes are emitted through the existing Postgres outbox and
  Valkey leaderboard worker.

The activity-time-tracking rollout is documented in
`../activity-time-tracking/README.md`. Migration
`0017_activity_time_tracking_schema` introduced nullable amount/unit data,
duration tracking, and `computed_score`.

## Compatibility invariants

The following behavior must remain true throughout the rollout:

- The feature gate remains off for ordinary users until rule-engine production
  verification is complete.
- The original form remains functionally unchanged.
- Existing clients may continue submitting `unit_id` UUIDs.
- Existing amount/unit payloads remain valid.
- Existing log and contest score snapshots never change merely because a new
  rule-set version is published.
- Historical reads and leaderboards continue using
  `coalesce(computed_score, score)` until the generated columns are removed.
- Amount-plus-duration logs continue scoring from amount.
- Duration is represented in seconds at the API boundary and converted to
  minutes only for rule evaluation.
- An unmatched platform or replacing-contest rule awards zero and still creates
  or updates the log successfully.
- Completed contest snapshots are never recalculated implicitly.
- Default tag suggestions do not carry scoring data. Tags affect scoring only
  when an explicit rule matches the normalized tag.
- Domain packages must not import storage packages.
- Time-dependent services use an injected `commondomain.Clock`.

## Stable-unit-key rollout

The stable-unit-key change is intentionally deployed in four steps: add nullable
columns, backfill existing rows, deploy writers that populate the keys, then
enforce the non-null and log-tracking constraints. The enforcement migration
must only run after every live API writer has the preceding application change;
otherwise a rolling deployment could reject writes from an old pod.

## Platform parity rule sets

The active platform rule set must reproduce established production scoring for
amount-and-unit inputs. The Writing output-activity boost was applied to the
production unit modifiers in 2023 but was not propagated to the original seed
migration or public manual. Platform rule-set v2 restores that compatibility;
shadow mode must reach zero unexplained differences before the rule set becomes
authoritative.

### Amount rules

| Activity | Stable unit key | Language | Rule kind | Rate |
| --- | --- | --- | --- | ---: |
| Reading | `reading_page` | any | base | 1 |
| Reading | `reading_two_column_page` | Japanese | base | 1.6 |
| Reading | `reading_comic_page` | any | base | 0.2 |
| Reading | `reading_sentence` | any | base | 0.05 |
| Reading | `reading_character` | Japanese, Korean, and Chinese variants | base | 0.0025 |
| Reading | `reading_character` | any | base | 0.000833333 |
| Listening | any | any | base | 0.4 |
| Listening | `listening_dense_minutes` | any | modifier | 1.5 |
| Writing | `writing_page` | any | base | 10 |
| Writing | `writing_sentence` | any | base | 0.5 |
| Writing | `writing_character` | Japanese, Korean, and Chinese variants | base | 0.025 |
| Writing | `writing_character` | any | base | 0.00833333 |
| Speaking | any | any | base | 0.5 |
| Speaking | `speaking_dense_minutes` | any | modifier | 1.4 |
| Study | `study_minute` | any | base | 0.5 |

The high-rate character language codes currently represented by database unit
rows are `jpn`, `kor`, `zho`, `cmn`, `yue`, and `wuu`. Their language-specific
rule is a more-specific non-stackable base; it wins before the broad character
base and is not modeled as a `3x` modifier.

### Duration rules

Duration-only scoring is a new extension beyond the amount-and-unit rules in
the public manual.

| Activity | Score source | Rate per minute |
| --- | --- | ---: |
| Reading | `duration_minutes` | 0.2 |
| Listening | `duration_minutes` | 0.4 |
| Writing | `duration_minutes` | 0.2 |
| Speaking | `duration_minutes` | 0.5 |
| Study | `duration_minutes` | 0.5 |

The high-density unit rules use relative stackable modifiers: Listening scores
at `0.4 * 1.5 = 0.6` per minute, and Speaking scores at
`0.5 * 1.4 = 0.7` per minute. A future normalized `dense` tag can use the same
modifiers without changing base-rate selection.

## Proposed persistence model

Use immutable published rule-set versions so historical provenance can safely
reference the rule that produced a score.

### `scoring_rule_sets`

- `id uuid primary key`
- `scope text not null` constrained to `platform` or `contest`
- `contest_id uuid null`
- `version integer not null`
- `status text not null` constrained to `draft` or `published`
- `mode text null` constrained to `override` or `replace`
- `fallback_rule_set_id uuid null`
- `created_at timestamp not null`
- `published_at timestamp null`

Platform sets have no contest, mode, or fallback. Contest sets belong to one
contest. An overriding contest set pins a published platform set through
`fallback_rule_set_id`; a replacing set has no fallback.

Published rule sets and their rules must not be updated or deleted. A change
creates a new version.

### `scoring_rules`

- `id uuid primary key`
- `rule_set_id uuid not null`
- `priority integer not null`
- `stackable boolean not null`
- `activity_id smallint not null`
- `unit_key text null`
- `language_code varchar(10) null`
- `tag varchar(50) null`
- `score_source text not null` constrained to `amount` or `duration_minutes`
- `rate real not null`

Priorities are unique within a rule set. Non-stackable rows are base rules;
stackable rows are modifiers. Tags are stored in the same normalized lowercase
form as log tags. Activity IDs are validated against the code-owned activity
catalog. Unit keys are validated against the code-owned unit catalog.

Use a singleton platform configuration row to point to the active published
platform rule set. Add a nullable `scoring_rule_set_id` to `contests` to pin a
published contest set.

### Unit compatibility

Add a stable key to the existing database unit rows as a legacy UUID-to-key
mapping. Multiple language-specific UUID rows may map to the same code-owned
key, such as `reading_character`.

Add `unit_key` to `logs`. Backfill it through the compatibility mapping. During
the transition:

- Old requests may send `unit_id`.
- New requests may send `unit_key`.
- The API rejects requests that send conflicting UUID and key values.
- New writes snapshot `unit_key`.
- UUID `unit_id` and the database unit rows remain available to old clients.
- The configuration API exposes stable keys additively before any field is
  removed.

### Score provenance

Add these fields to both `logs` and `contest_logs`:

- `score_rule_set_id`
- `score_rule_ids`
- `score_rates`
- `score_source`

`computed_score` remains the authoritative numeric snapshot. For a matched
rule set, the ordered rule ID and rate arrays contain the selected base rule
first, followed by every applied modifier in ascending priority order.
Non-selected matching base rules are omitted. For an unmatched rule set,
`computed_score` is zero, `score_source` records the selected input source, and
the rule-set/rules/rates fields may be null. Historical rows may have all
provenance fields null.

## Rule resolution paths

### Create log

1. Validate authentication, tags, activity, tracking data, and registrations.
2. Resolve a stable unit key from either the new key or legacy UUID.
3. Resolve the active platform rule set and calculate the base score.
4. Resolve each selected contest independently.
5. Persist the base log, contest snapshots, tags, and leaderboard outbox events
   in one transaction.

### Update log

1. Load the existing log and check ownership or admin access.
2. Resolve the platform score using the current active platform set.
3. Resolve each ongoing attached contest using its pinned rules.
4. Update the base log and ongoing contest snapshots in one transaction.
5. Leave completed contest snapshots unchanged.

### Attach an existing log to a contest

Resolve the contest score at attachment time from the log's snapshotted tracking
inputs and the contest's pinned rules. Do not copy the base log score.

### Preview

The preview endpoint and write commands must call the same domain evaluator and
rule-resolution service. Preview is advisory; create/update always resolves
again inside the authoritative backend flow.

## Repository touchpoints

### Domain

- `services/immersion-api/domain/activities.go`
- `services/immersion-api/domain/models.go`
- `services/immersion-api/domain/logtracking.go`
- `services/immersion-api/domain/logcreate.go`
- `services/immersion-api/domain/logupdate.go`
- `services/immersion-api/domain/logcontestupdate.go`

Add scoring types and the pure evaluator in focused domain files. Define narrow
rule-loading interfaces at their consumers.

### Postgres

- `services/immersion-api/storage/postgres/migrations/`
- `services/immersion-api/storage/postgres/queries/units.sql`
- `services/immersion-api/storage/postgres/queries/logs.sql`
- `services/immersion-api/storage/postgres/queries/contests.sql`
- `services/immersion-api/storage/postgres/queries/leaderboard.sql`
- `services/immersion-api/storage/postgres/queries/contest_profile.sql`
- `services/immersion-api/storage/postgres/repository/repo_createlog.go`
- `services/immersion-api/storage/postgres/repository/repo_updatelog.go`
- `services/immersion-api/storage/postgres/repository/repo_updatelogcontests.go`

SQL keywords in new migrations and queries must remain lowercase. Regenerate
sqlc and Gazelle metadata after query, import, or file changes.

### HTTP and frontend

- `services/immersion-api/http/rest/openapi/api.yaml`
- `frontend/apps/webv2/app/immersion/api.ts`
- `frontend/apps/webv2/app/immersion/NewLogForm/domain.tsx`
- `frontend/apps/webv2/app/immersion/NewLogForm/Form.tsx`
- `frontend/apps/webv2/app/immersion/NewLogFormV2/domain.tsx`
- `frontend/apps/webv2/app/immersion/NewLogFormV2/Form.tsx`
- `frontend/apps/webv2/app/immersion/EditLogForm/Form.tsx`
- `frontend/apps/webv2/app/immersion/ContestForm.tsx`

Keep deprecated modifier fields until all frontend estimates use the preview
endpoint. All form changes must use `react-hook-form` and components from the
`ui` package.

## Trunk-based rollout plan

- [x] Document the rule evaluation semantics in ADR 005.
  - [x] Rules use explicit priority ordering.
  - [x] Matchers use activity, stable unit key, language, and one normalized tag.
  - [x] Rules declare `score_source: amount | duration_minutes`.
  - [x] Amount-plus-duration logs use `amount`.
  - [x] Unmatched rules produce zero.
  - [x] Published rule sets are immutable and versioned.

- [x] Introduce stable, code-owned unit keys.
  - [x] Add keys such as `reading_page`, `reading_character`, and
    `listening_minute`.
  - [x] Add `unit_key` to logs and API configuration responses.
  - [x] Backfill existing logs from their UUID units.
  - [x] Continue accepting legacy `unit_id` UUID payloads.
  - [x] Keep the legacy form's behavior unchanged.

- [x] Add scoring-rule storage.
  - [x] Add versioned `scoring_rule_sets`.
  - [x] Add ordered `scoring_rules`.
  - [x] Support platform-owned and contest-owned rule sets.
  - [x] Add `override` and `replace` contest modes.
  - [x] Pin overriding contest sets to a specific platform rule-set version.
  - [x] Seed a platform rule set that reproduces current behavior.

- [x] Implement the domain scoring engine.
  - [x] Accept activity, unit key, language, normalized tags, amount, and duration.
  - [x] Select the log's score source.
  - [x] Select the first matching non-stackable base rule in priority order.
  - [x] Multiply every matching stackable modifier into the base score.
  - [x] Return the selected base then applied modifiers as provenance.
  - [x] Return zero with no applied rule when nothing matches.
  - [x] Keep domain interfaces narrow and storage-independent.

- [x] Add comprehensive engine tests.
  - [x] Cover specific base selection, omitted fallback bases, and priority ordering.
  - [x] Cover multiple matching modifiers and modifiers without a base.
  - [x] Cover activity, unit, language, and tag matching.
  - [x] Cover amount and duration score sources.
  - [x] Cover amount-plus-duration precedence.
  - [x] Cover override fallback to platform rules.
  - [x] Cover replace mode producing zero for uncovered inputs.
  - [x] Cover platform rules producing zero for uncovered inputs.
  - [x] Cover stable behavior when published rules are superseded.

- [ ] Run the engine in shadow mode.
  - [x] Evaluate both the interim scorer and the rule engine.
  - [x] Keep writing the interim result.
  - [x] Record mismatches and unmatched inputs.
  - [ ] Verify all existing amount/unit combinations.
  - [x] Use `0.4` per minute for duration-only Listening.
  - [x] Do not introduce dense-tag scoring in this plan.

- [x] Add score provenance snapshots.
  - [x] Keep `computed_score` as the authoritative score.
  - [x] Add the applied rule-set ID and ordered rule IDs.
  - [x] Add the ordered applied rates.
  - [x] Add `score_source`.
  - [x] Add the same fields independently to `contest_logs`.
  - [x] Allow rule fields to be null when the score is zero because no rule matched.

- [x] Switch platform log scoring to the engine behind the feature flag.
  - [x] Use the active platform rule set for create and update.
  - [x] Preserve legacy UUID-unit compatibility.
  - [x] Preserve existing amount-plus-duration behavior.
  - [x] Confirm historical logs and aggregates remain unchanged.

- [x] Implement independent contest scoring.
  - [x] Resolve each selected contest's rules separately.
  - [x] Snapshot each contest score in `contest_logs`.
  - [x] Re-resolve ongoing contest scores when a log is edited.
  - [x] Resolve rules when an existing log is attached to a contest.
  - [x] Never change completed-contest snapshots implicitly.
  - [x] Confirm uncovered replacing rule sets award zero.

- [x] Add a score-preview API.
  - [x] Accept the same activity, unit, language, tags, amount, duration, and contest
    selection as log submission.
  - [x] Return the platform estimate and applicable contest estimates.
  - [x] Return applied-rule information where appropriate.
  - [x] Return zero for uncovered inputs.

- [x] Move frontend estimates to score preview.
  - [x] Use `react-hook-form` values for preview requests.
  - [x] Remove dependence on `unit.modifier`.
  - [x] Preserve the legacy form's visible workflow.
  - [x] Display platform and contest-specific estimates clearly.
  - [x] Keep the feature behind the existing flag until verified.

- [x] Add rule-set management APIs.
  - [x] Create draft rule-set versions.
  - [x] Validate priorities, matchers, score sources, and rates.
  - [x] Publish immutable versions.
  - [x] Activate platform versions.
  - [x] Configure contest mode and pinned fallback.
  - [x] Prevent implicit mutation of published rule sets.

- [x] Add contest scoring configuration UI.
  - [x] Use components from the `ui` package.
  - [x] Use `react-hook-form` for form state and validation.
  - [x] Support ordered rules and match conditions.
  - [x] Show uncovered inputs as scoring zero.
  - [x] Show whether the contest overrides or replaces platform rules.
  - [x] Prevent unsafe rule changes after the contest starts.

- [ ] Complete production verification.
  - [ ] Shadow mismatches are understood or zero.
  - [ ] Legacy amount logs still score correctly.
  - [ ] Duration-only Listening uses `0.4`.
  - [ ] Uncovered contest inputs score zero.
  - [ ] Platform and contest snapshots aggregate correctly.
  - [ ] Old clients and the legacy form remain functional.

- [ ] Remove the interim scoring bridge.
  - [ ] Delete `ComputeInterimLogScore`.
  - [ ] Stop reading rates from database unit modifiers.
  - [ ] Remove deprecated modifier API fields after client migration.
  - [ ] Remove the generated legacy `score` columns when all reads use
    `computed_score`.
  - [ ] Remove obsolete database-owned unit metadata after UUID compatibility ends.

## Completion and validation checklist

A rollout item is complete only when its implementation subitems are checked,
its relevant tests pass, generated files are current, and the change remains
independently deployable.

- [x] Format changed backend Go files.

  ```sh
  gofmt -w services/
  ```

- [x] Regenerate sqlc after changing Postgres queries.

  ```sh
  cd services/immersion-api/storage/postgres
  go generate
  ```

- [x] Regenerate OpenAPI code after changing the API specification.

  ```sh
  cd services/immersion-api/http/rest/openapi
  go generate
  ```

- [x] Regenerate Bazel metadata after adding files or changing Go dependencies.

  ```sh
  bazel run //:gazelle
  ```

- [x] Run focused domain, repository, and REST tests for the rollout slice.

  ```sh
  bazel test //services/immersion-api/domain:domain_test
  bazel test //services/immersion-api/storage/postgres/repository:repository_test
  bazel test //services/immersion-api/http/rest:rest_test
  ```

- [x] Build and test all backend services before merging a completed slice.

  ```sh
  bazel build //services/...
  bazel test //services/...
  ```

- [x] Typecheck and lint WebV2 after frontend or generated API changes.

  ```sh
  cd frontend
  pnpm --filter webv2 exec tsc --noEmit
  pnpm --filter webv2 lint
  ```

- [x] Build the frontend before merging a user-visible rollout slice.

  ```sh
  cd frontend
  pnpm build
  ```

- [x] Confirm generated files and build metadata have no unexplained diff.

  ```sh
  git diff --check
  bazel run //:gazelle -- -mode=diff
  ```

- [ ] Verify shadow mode before switching authoritative scoring.
  - [ ] The ordinary-user feature gate remains disabled.
  - [ ] Legacy amount/unit requests still succeed.
  - [ ] Existing unit/language combinations have zero unexplained mismatches.
  - [ ] Duration rules match the documented parity rates.
  - [ ] Unmatched inputs resolve to zero without rejecting writes.
  - [ ] Shadow diagnostics contain no descriptions, tags, user identifiers, or
    other user content.

- [ ] Verify production after switching authoritative scoring.
  - [ ] Historical effective scores and leaderboard totals remain unchanged.
  - [ ] New amount-only logs contain valid score snapshots and provenance.
  - [ ] New duration-only logs contain valid score snapshots and provenance.
  - [ ] New amount-plus-duration logs score from amount.
  - [ ] Base and contest snapshots use their intended rule sets.
  - [ ] Replacing contest sets award zero for uncovered inputs.
  - [ ] Overriding contest sets use their pinned platform fallback.
  - [ ] Completed contest snapshots remain unchanged after log edits.
  - [ ] Outbox-driven leaderboard refreshes converge to Postgres totals.

## Follow-up

Once the scoring engine is working, write a separate migration plan to remove
`listening_dense_minutes`. Dense scoring and that unit's data migration are
intentionally outside this plan.
