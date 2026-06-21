---
sidebar_position: 4
title: Activity Time Tracking
---

# Activity time tracking

## Status

This document tracks the activity time-tracking work for trunk-based development on `main`.

The change should be built as a series of small, deployable commits merged directly into `main`. Each step should be backward compatible with the currently deployed frontend and database schema. The original logging form should keep behaving as it does today; the first user-visible time-tracking experience belongs in the V2 logging form.

## Goals

- Add first-class time tracking to logs with a duration field.
- Keep the original amount/unit logging workflow working for existing users.
- Let the V2 logging form support time-first activities.
- Avoid bringing back database-owned activities or database-owned default tags.
- Keep the design compatible with [ADR 005: Scoring rules](../../docs/adr/005-scoring-rules.md), even though rules-based scoring is not part of this task.
- Make each backend and frontend change safe to deploy independently.

## Non-goals

- Do not implement database-backed scoring rules in this task.
- Do not migrate activity IDs to string keys in this task.
- Do not make units code-owned in this task.
- Do not remove legacy amount/unit scoring while the original logging form still depends on it.
- Do not add scoring metadata to default tags.

## Current architecture constraints

Activities are code-owned domain values. The `log_activities` table is no longer a runtime dependency and should not be reintroduced for activity metadata such as input type or time scoring.

Default tags are code-owned suggestions. Logged tags remain free text. Tags can affect scoring later only when a scoring rule explicitly matches them.

Units still exist in the database for the current amount/unit logging flow. They are still used to validate amount logs and to snapshot the legacy `modifier`. ADR 005 says units should eventually become code-owned stable identifiers, but that is a later migration.

Scoring currently uses legacy snapshots: `amount`, `modifier`, and generated `score`. ADR 005 changes this later to ordered database-backed scoring rules with score snapshots. Time tracking should prepare for that by separating "what the user tracked" from "how score was calculated".

## Activity input model

The activity set stays fixed in code:

| ID | Activity | V2 input mode |
| --- | --- | --- |
| 1 | Reading | amount primary, optional time |
| 2 | Listening | time primary |
| 3 | Writing | amount primary, optional time |
| 4 | Speaking | time primary |
| 5 | Study | time primary |

The input mode is UI metadata owned by domain code. It should not be stored in Postgres. The API can expose it as an additive field on activity configuration responses, for example `input_type: "amount_primary" | "time_primary"`.

The backend should not reject amount/unit logs for time-primary activities yet. The original logging form may still send "30 minutes" as amount/unit for Listening, Speaking, or Study. V2 can hide amount/unit for those activities while the backend remains backward compatible.

## Data model

Add `duration_seconds` to `logs` as nullable metadata.

Keep `amount`, `unit_id`, and `modifier` for the legacy amount/unit path, but allow them to be nullable so V2 can create duration-only logs.

Add an application-written score snapshot column, `computed_score`, for logs that cannot use the generated `amount * modifier` column. This is the long-term score snapshot column. Existing logs keep using the generated `score` during the transition, but future code should write scores into `computed_score`, and a later cleanup should delete the old generated `score` column.

Read paths should use an effective score:

```sql
coalesce(computed_score, score)
```

The same model applies to `contest_logs`: keep the existing generated `score` for historical amount logs, add nullable `duration_seconds` and `computed_score`, and aggregate with the same effective-score expression.

Database constraints should require at least one complete tracking method:

```sql
duration_seconds is not null
or (amount is not null and unit_id is not null and modifier is not null)
```

Because generated `score` can become null for duration-only rows, the migration must also make the generated score columns nullable where needed.

## Interim scoring before ADR 005

Rules-based scoring is next, not part of this task.

Until then, use a small legacy scoring bridge in domain code:

- Amount/unit submissions keep the current behavior: score is based on the selected unit modifier.
- If both amount/unit and duration are present, amount/unit remains the scoring source and duration is metadata.
- Duration-only Listening submissions use a temporary `minutes * 0.4` fallback, matching the plain minute row in the manual.
- Duration-only Speaking and Study submissions use a temporary `minutes * 0.5` fallback, matching the plain minute row in the manual.
- Duration-only Reading and Writing submissions use a temporary `minutes * 0.2` fallback.

The interim time scoring policy should be isolated in one domain helper with a clear ADR 005 TODO. Do not attach a `time_modifier` to `Activity`, do not add it to a database activity table, and do not spread time scoring constants through handlers or repositories. The duration-first V2 UI must not ship until ADR 005 scoring rules can preserve dense-minute behavior for Listening and Speaking.

When ADR 005 scoring rules are implemented, this bridge should be replaced by the rule engine. New logs should still snapshot the resolved score into `computed_score`.

## API behavior

Create/update log requests should accept `duration_seconds` as an optional field. The API boundary uses seconds. The frontend should let users enter minutes in the form, then convert to seconds before submitting.

For backward compatibility:

- Existing amount/unit payloads remain valid.
- Existing responses keep returning amount, unit, modifier, score, and activity IDs/names.
- New response fields are additive.
- V2 can prefer `duration_seconds` for display when present.

Validation:

- `duration_seconds`, when present, must be positive.
- Amount/unit logs must still include a valid unit for the selected activity.
- Duration-only logs are valid for every activity. Listening, Speaking, and Study use the legacy minute-unit emulation. Reading and Writing use the fallback duration formula.
- Unknown activity IDs still fail through the code-owned activity validation path.
- Contest attachment validation should use the log activity ID as it does today; adding duration must not bypass contest allow-lists.

## Frontend behavior

Original logging form:

- Keep the existing amount/unit flow.
- Do not force users into the new time-tracking UI.
- It may ignore `input_type` and `duration_seconds`.

V2 logging form:

- Reading and Writing show amount/unit as the primary input.
- Reading and Writing may expose optional time spent as expandable metadata.
- Listening, Speaking, and Study show time spent as the primary input.
- Users enter duration in minutes; V2 submits `duration_seconds` to the API.
- Activity switching resets invalid hidden fields before submit.
- Edit and detail views display duration when it exists.

## Trunk-based rollout plan

Each step should be merged to `main` and deployable on its own.

1. [x] Backend schema: add nullable duration/effective-score support and constraints without changing API behavior.
2. [x] Backend reads: switch log, contest log, leaderboard, and yearly split queries to effective score.
3. [x] Backend writes: accept optional `duration_seconds` and compute `computed_score` for duration-only logs.
4. [ ] Backend config: expose code-owned activity input mode as additive activity metadata.
5. [ ] Backend tests: cover legacy amount logs, amount+duration logs, duration-only time-primary logs, duration-only Reading/Writing fallback scoring, contest attachment, and leaderboard aggregation.
6. [ ] Frontend API types: add optional duration and activity input mode.
7. [ ] V2 create/edit forms: add adaptive time-tracking UI.
8. [ ] V2 detail/list views: display duration when present.
9. [ ] Original form regression check: create and edit logs through the existing amount/unit flow.
10. [ ] Production rollout check: confirm old logs, new amount logs, and new duration logs aggregate correctly.
11. [ ] Follow-up: implement ADR 005 scoring rules and replace the interim scoring bridge.

## Resolved decisions

- Duration-only Listening, Speaking, and Study should emulate the existing minute-based amount/unit scoring until rules-based scoring exists.
- V2 should allow duration-only Reading and Writing immediately, using a fallback scoring formula.
- `computed_score` is the long-term score snapshot column. Scores should be calculated in application code instead of a database-generated formula, and the old generated `score` column should be deleted in a future cleanup.
- The API should accept duration in seconds. The frontend should let users enter minutes and convert to seconds before submit.
