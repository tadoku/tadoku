# Brainstorm: Activity Time Tracking + Optional Amount/Unit

## Context

Currently every log entry requires an amount + unit (e.g. "50 pages", "30 minutes"). The goal is to let any activity track time spent as a first-class field, make amount/unit optional, and adapt the form based on the activity type so users see the fewest fields necessary.

## Chosen Direction

**Form**: Flow A — Adaptive form with manual time entry (no timer in v1)
**Scoring**: Time-based scoring for Listening/Speaking/Study; amount-based for Reading/Writing with a low fallback when only time is provided

---

## Part 1: Form UX — Adaptive Form

The form adapts its layout based on the selected activity. The activity determines which field is prominent and which is expandable.

### Reading / Writing selected

Amount is the primary input. Time is optional and collapsed.

```
Language:    [Japanese          ▼]
Activity:    [Reading           ▼]

Amount:      [  50  ] [Pages   ▼]       ← primary, always visible
  + Track time spent                     ← link, expands to time input

Description: [One Piece vol 45    ]
Tags:        [Manga] [Fiction]
                          Est. score: 50.0
```

After expanding time:
```
Amount:      [  50  ] [Pages   ▼]
Time spent:  [  30  ] min    [×]        ← closable
```

**Validation for Reading/Writing:**
- Amount + unit is required (keeps scoring precise)
- Time is optional metadata
- If we want to allow time-only Reading/Writing: amount becomes optional too, score uses a low fallback modifier (e.g. 0.3 pts/min vs 1.0 pts/page)

### Listening / Speaking / Study selected

Time is the **only** input. No amount/unit fields at all.

```
Language:    [Japanese          ▼]
Activity:    [Listening         ▼]

Time spent:  [  30  ] min               ← only tracking input

Description: [Podcast episode 42  ]
Tags:        [Podcast]
                          Est. score: 9.0
```

**Validation for Listening/Speaking/Study:**
- Time is required
- No amount/unit fields shown or accepted
- Score: `(minutes) * activity.time_modifier` (e.g. 0.3 for listening)

### Activity switching behavior

When the user changes activity, the form adapts:
- Switching from Reading → Listening: amount+unit fields disappear entirely, time field appears
- Switching from Listening → Reading: time field disappears, amount+unit appears (time available via expand link)
- Any expanded optional time field collapses on activity switch
- Unit selector resets to first available unit for the new activity (existing behavior)

### Key implementation details

**Files to modify (frontend):**
- `frontend/apps/webv2/app/immersion/NewLogFormV2/Form.tsx` — main form layout, conditional rendering based on activity
- `frontend/apps/webv2/app/immersion/NewLogFormV2/domain.tsx` — schema changes (amountValue optional, add timeMinutes field)
- `frontend/apps/webv2/app/immersion/NewLogForm/Form.tsx` — contest-aware form, same changes
- `frontend/apps/webv2/app/immersion/EditLogForm/Form.tsx` — edit form
- `frontend/apps/webv2/app/immersion/api.ts` — API payload types, `Activity` type gains `input_type: "time" | "amount"`
- `frontend/packages/ui/components/Form/AmountWithUnit.tsx` — may need to support optional state

**New Zod schema (sketch):**
```typescript
const schema = z.object({
  languageCode: z.string().length(3),
  activityId: z.number(),
  timeMinutes: z.number().positive().optional(),  // NEW
  amountValue: z.number().positive().optional(),   // was required
  amountUnit: z.string().optional(),               // was required
  tags: z.array(z.string().max(50)).max(10),
  description: z.string().optional(),
}).superRefine((data, ctx) => {
  const activity = data.activities?.find(a => a.id === data.activityId)
  if (activity?.input_type === 'time') {
    if (!data.timeMinutes) ctx.addIssue({ code: 'custom', path: ['timeMinutes'], message: 'Time is required' })
  } else {
    // "amount" input type: need at least time or amount
    if (!data.timeMinutes && !data.amountValue)
      ctx.addIssue({ code: 'custom', path: ['amountValue'], message: 'Enter time or amount' })
  }
})
```

The frontend determines the input mode from the activity config (see below).

---

## Part 2: Database Storage

### Migration

```sql
-- Add duration column
alter table logs
  add column duration_seconds int;

-- Make amount/unit/modifier optional
alter table logs
  alter column amount drop not null,
  alter column unit_id drop not null,
  alter column modifier drop not null;

-- Ensure at least one complete tracking method:
-- either duration, or the full amount triple (amount + unit_id + modifier)
alter table logs
  add constraint logs_has_tracking_data
  check (
    duration_seconds is not null
    or (amount is not null and unit_id is not null and modifier is not null)
  );

-- Add time-based modifier and input type to activities
alter table log_activities
  add column time_modifier real not null default 0.3,
  add column input_type varchar(10) not null default 'amount';
  -- 'time' = Listening, Speaking, Study (time-only input)
  -- 'amount' = Reading, Writing (amount+unit primary, time optional)

update log_activities set input_type = 'time' where id in (2, 4, 5);
update log_activities set input_type = 'amount' where id in (1, 3);

-- Keep existing generated `score` column untouched (preserves all historical scores).
-- Add a new `computed_score` column for application-written scores (used for new logs).
alter table logs
  add column computed_score real;

-- contest_logs: same treatment (has amount, modifier, and generated score from migration 0013)
alter table contest_logs
  add column duration_seconds int;
alter table contest_logs
  alter column amount drop not null,
  alter column modifier drop not null;
-- Keep existing generated `score` (amount * modifier) intact for historical contest entries.
-- Add computed_score for new entries.
alter table contest_logs
  add column computed_score real;
```

### Score columns: dual-column approach

**`score`** (existing) — generated column: `amount * modifier`. Stays forever. For existing logs with amount/unit, this continues to work. For new time-only logs (where amount/modifier are NULL), this generates NULL or 0.

**`computed_score`** (new) — regular column written by the application at insert/update time. For new logs, this is the authoritative score. For old logs, this is NULL (the app reads `score` as fallback).

**Read logic** (Go):
```go
func (log *Log) EffectiveScore() float32 {
    if log.ComputedScore != nil {
        return *log.ComputedScore
    }
    return log.Score // fallback to generated column
}
```

All leaderboard/aggregation queries use `coalesce(computed_score, score)` to read the effective score.

### Score computation (Go, at insert/update time)

```
activity = lookupActivity(req.ActivityID)

if activity.InputType == "time":
    // Listening, Speaking, Study — score always from time
    computed_score = (duration_seconds / 60) * activity.TimeModifier
else if activity.InputType == "amount":
    if amount != nil && modifier != nil:
        computed_score = amount * modifier                              // primary: amount-based
    else if duration_seconds != nil:
        computed_score = (duration_seconds / 60) * activity.TimeModifier  // fallback: time-based (lower modifier)
```

For `input_type = "time"` activities, the backend rejects requests that include amount/unit (they shouldn't be sent).

For `input_type = "amount"` activities (Reading/Writing), when both time and amount are provided, amount-based scoring wins (more precise). When only time is provided, score uses a lower `time_modifier` (e.g. 0.3) to incentivize precise amount tracking while still giving some credit.

**Optional future cleanup**: A background migration could backfill `computed_score` for all old logs (`update logs set computed_score = score where computed_score is null`), after which all code could read from `computed_score` only.

### Domain model changes

**Files to modify (backend):**
- `services/immersion-api/domain/models.go` — `Log` struct: add `DurationSeconds *int`, `ComputedScore *float32`, and `EffectiveScore()` method. Make `Amount`/`UnitID`/`Modifier` pointers. `Activity` struct: add `TimeModifier float32`, `InputType string`.
- `services/immersion-api/domain/logcreate.go` — `LogCreateRequest`: add `DurationSeconds`, make `Amount`/`UnitID` optional. Score computation logic.
- `services/immersion-api/domain/logupdate.go` — same changes for update path
- `services/immersion-api/storage/postgres/queries/logs.sql` — update INSERT/SELECT queries, use `coalesce(computed_score, score)` in all score reads
- `services/immersion-api/storage/postgres/queries/activities.sql` — include `time_modifier`, `input_type`
- Leaderboard/aggregation queries — replace `score` with `coalesce(computed_score, score)` everywhere (both `logs` and `contest_logs` tables)
- `services/immersion-api/http/rest/openapi/api.yaml` — API spec changes: Activity response includes `input_type`, log create/update accepts optional `duration_seconds`
- New migration file in `services/immersion-api/storage/postgres/migrations/`

### Existing data

No backfill needed. All existing logs retain their amount/unit/modifier/score. `duration_seconds` will be NULL for historical entries.

---

## Part 3: Scoring Summary

| Activity | Primary input | Score source | Time modifier |
|----------|--------------|-------------|--------------|
| Reading | Amount + Unit | `amount * modifier` | 0.3 (fallback only) |
| Listening | Time | `minutes * 0.3` | 0.3 |
| Writing | Amount + Unit | `amount * modifier` | 0.3 (fallback only) |
| Speaking | Time | `minutes * 0.3` | 0.3 |
| Study | Time | `minutes * 0.3` | 0.3 |

When both time AND amount are provided → score from amount (more precise).
When only time is provided for Reading/Writing → low fallback score.

---

## Verification

1. **Backend**: `bazel test //services/immersion-api/...` — all existing tests pass + new tests for time-only log creation, score computation, nullable amount fields
2. **Frontend**: `pnpm --filter webv2 exec tsc --noEmit` + `pnpm --filter webv2 lint` — type checks pass with new optional fields
3. **Manual**: Create logs in each mode (time-only, amount-only, both) for each activity type and verify scores
4. **Migration**: Run migration against a test database, verify existing logs are unchanged
