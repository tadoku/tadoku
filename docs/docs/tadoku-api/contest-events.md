---
title: Contest event feeds
description: Poll contest submission events with event-ID cursors and fetch current contest log contents for previews.
---

# Contest event feeds

Read this when building an integration that follows a contest's submissions.

A contest feed reports changes to that contest's logs. Each event contains only
`id`, `log_id`, `event` and `occurred_at`. Fetch current log contents separately when an
integration needs a preview or the contest's score.

## Read events

```http
GET /api/internal/immersion/contests/{id}/events?cursor=0&limit=100
```

| Event | Meaning |
| --- | --- |
| `log.submitted` | A log starts appearing in this contest. |
| `log.updated` | A submitted log's contest-visible fields change. |
| `log.withdrawn` | A log leaves this contest, including deletion or moderation. |

Creating a log before submitting it produces no event for the contest. Creating
and submitting together produces one submitted event. Editing after withdrawal
produces no event there. Submitting the log again produces another submitted
event. Removing registered languages produces withdrawals for the affected logs.
Deleting an account withdraws its logs from unfinished contests and updates its
logs kept in finished contests, whose description and tags it clears.

Updates describe the fields visible in the contest. A personal score change can
leave a closed contest's score frozen and produce no event there; shared
content changes such as description or tags still produce updates. No-op and
failed writes produce no events. Renaming users or activity labels and
recalculating a leaderboard are outside this feed.

The response contains `events`, `next_cursor` and `has_more`. Events arrive in
occurrence order, oldest first, with the event UUID breaking equal-timestamp
ties. `occurred_at` is stored once using the source operation's timestamp; it
does not change when the feed is read. The server resolves an event-ID cursor
to that event's stored `(occurred_at, id)` pair. Use the response order.

Occurrence order is not commit order. If a transaction commits after a consumer
has advanced past its occurrence time, its event can appear before the saved
cursor. Forward polling alone does not guarantee delivery of those late events.
Replaying earlier history and deduplicating by event ID can recover them; there
is no bounded replay window that guarantees recovery of every delayed commit.

- Omit `cursor`, or pass `0`, to start at the earliest recorded event.
- Pass `now` once to receive an empty page with the latest visible event ID,
  or `0` if the feed is empty.
- Pass an event ID to read strictly after that event.
- `limit` defaults to 100 and accepts integers from 1 through 100.

A nonempty page's `next_cursor` equals its final event's ID. An ordinary empty
page preserves the input event ID, or returns `0` when starting from the
beginning. Save this cursor after processing the page succeeds. Fetch the next
page immediately while `has_more` is true; otherwise poll later with the saved
cursor. Do not keep requesting `now`, which would skip intervening changes.

Retries can return the same event. Its stable ID supports deduplication, but the
API does not guarantee exactly-once external side effects. A cursor is scoped
to its contest and tenant. Malformed, unknown or wrong-scope IDs return 400.

## Enrich previews with current contents

```http
GET /api/internal/immersion/contests/{id}/logs/{logId}
```

Use the contest ID from the feed URL and the log ID from the event. This read
returns the ordinary contest log representation, including the contest's score
and tracking values. It omits other contest registrations. The generic personal
log endpoint is not a substitute because its score can differ.

A board consumer can refresh a log for any of the three events:

1. Fetch the current contest log.
2. On 200, create or replace the preview with the returned contents.
3. On 404, remove the preview or mark it withdrawn.
4. On a retryable failure, keep the saved cursor and retry. Do not interpret an
   authentication failure as a withdrawal.
5. Save the page cursor after every event has been processed successfully.

Process refreshes for one log in order. A log may change again or be withdrawn
before its content fetch. It may also be resubmitted before a withdrawal event
is processed. The content endpoint returns current state, not the content at
the event's occurrence time. No log contents are embedded in event records.

## Access and history

Both reads use the existing ordinary contest-log access rules and the shared
authentication and ban pipeline. Public reads work anonymously. There is no
feed registration, opt-in or separate consent flow. A cursor grants no extra
access. Responses use `Cache-Control: private, no-store`.

Recorded events remain available for the lifetime of the contest, including
after a submission is withdrawn or its source log is deleted. Deleting the
contest makes its feed unavailable. Recording begins when event publication is
enabled; older logs do not receive fabricated historical submitted events.

The feed describes recorded changes rather than a complete initial snapshot.
The existing page-number contest log list can shift under concurrent changes,
so enumerating it and then replaying events does not guarantee a complete
initial replica. Personal feeds, historical content snapshots and a consistent
full-state bootstrap are outside this API.
