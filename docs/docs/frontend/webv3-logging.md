---
title: Webv3 activity logging
description: Native Webv3 activity creation, tags and their suggestions, remembered choices, saved detail, edits, deletion, contest submissions and the score estimate, their draft owners, session safety and browser journeys.
sidebar_position: 3
---

# Webv3 activity logging

Read this when you change a Webv3 log page, the log form or its data.

`/logs/new` prompts anonymous visitors to sign in through authv2, then loads
configuration and ongoing contest registrations from the selected tenant. The
form follows Webv2's: the language list starts with the languages the viewer
used before, then lists every language. Language and activity determine the
amount and unit input; Time spent, in minutes, is required for timed activities
and optional for the others. Changing the activity clears the quantity, because
activities count in different units. Changing the language keeps the unit when
the new language offers it, so the quantity keeps its meaning. A missing unit is reported on the amount
field, which shows one error for the amount and its unit. The description is one
line. As in Webv2, `/logs/new?amount=<n>` prefills a positive whole amount; any
other value is ignored. Cancel returns home, and Log activity saves.

Tags are chips, at most 10. While the viewer types, and once typing pauses for
300 ms, the tag suggestion operation offers the viewer's own tags containing the
text with their usage counts, such as `manga (3×)`, followed by common tags; a
failed read suggests nothing. Beside the form on wide screens, and below the
tags on narrow ones, the activity's score modifier and common tags toggle in
the draft, a chosen one struck through; adding stops at 10 tags. Paper's `TagsInput` reports the typed text
through `onQueryChange` and labels suggestions with `formatOption`.

After a save the browser remembers the log's language and activity under
Webv2's `log-form-v2:last-logged-language` and `log-form-v2:last-logged-activity`
keys, so they carry over between the two hosts. Once the page has hydrated, a
new draft starts from them when they are still offered; a remembered timed
activity starts without an amount, as in Webv2. Without them, the draft starts
with the viewer's first previously used language and the first activity.
Unavailable browser storage is ignored.

Contests are not chosen in the form. When the viewer has ongoing registrations,
a save opens the saved log's contest step at `/logs/$logId/contests?preselect=1`,
as Webv2 does; otherwise it opens the saved log. A log whose contest step is
left unsaved stays in the personal record only.

The draft has three owners:

- `frontend/apps/webv3/src/lib/log-draft.ts` owns the form's values: their
  defaults for a new or saved log, including remembered choices, the unit that
  follows a language or activity change, validation with the API's rules, and
  conversion to the `NewLog`, `LogUpdate` and score preview inputs the server
  functions accept. `frontend/apps/webv3/src/lib/log-preferences.ts` reads and
  stores the remembered choices.
- `frontend/apps/webv3/src/components/LogForm.tsx` composes the new and edit
  screens from `usePaperForm` and Paper's bound controls, and owns navigation,
  tag suggestions, the score estimate and failure messages.
  `frontend/apps/webv3/src/components/TagsSidebar.tsx` holds Webv2's
  per-activity tag lists.
- `frontend/apps/webv3/src/lib/logs.ts` owns the server functions, their query
  keys and `useSaveLog()`, `useUpdateLog()`, `useUpdateLogContests()` and
  `useDeleteLog()`, which cache a saved log for its detail page and refresh the
  viewer after an account switch. Each marks the reads its change affects as
  stale by query-key prefix, keyed by the log's owner rather than the account
  making the change, so the next page to show one reads it again without a
  reload. Another browser's cache can't be reached this way: it catches up when
  its copy goes stale.

`frontend/apps/webv3/src/components/DraftOwnerGate.tsx` fixes the account that
owns a page's draft, for both forms.

`frontend/apps/webv3/src/lib/logs.server.ts` calls the configuration,
registrations, tag suggestion, create, detail, update, delete and score preview operations
through the generated client, which validates their responses, with the session
cookies described in [account sessions](webv3.md#account-sessions). A registration without its ID or contest makes the
options unavailable. The API owns scoring, contest eligibility, authorization
and persistence; Webv3 never calculates a score itself.

Both forms show the API's estimated score in the footer of the form card,
"Estimated score: X", once the draft is valid and has not changed for 300 ms. The request is a GET server function and sends no contest
registrations, so it estimates the personal score only. An incomplete draft, a
pending first estimate and a failed one show an em dash; a zero estimate shows N/A.
A saved amount names its unit by the configuration's `unit_key`; the API resolves
the language's variant.

The saved log's detail at `/logs/$logId` displays the persisted language,
measure, tags, score and contests. Reload reads that
record again. An expired session cookie retries a public detail read
anonymously. Public detail omits private contest submissions; only the owner
and administrators get the link to manage them.

Administrators manage any learner's log as its owner does: the detail page
offers them Edit log, Delete log and the contest submissions, and the edit and
contest pages admit them. The pages decide this from the viewer's role, and the
API, which accepts a change from the log's owner or an administrator, stays the
authority. `DraftOwnerGate` records whether the draft's account was an
administrator when the draft started. Their pages name another learner's log in the third person, such as This log
and the learner's personal record, and leave out the advice to join a contest
or change a registration, which concerns the viewer's own contests.

The detail page's heading is the log's description when it is nonempty, has at
most 60 characters and has no NSFW tag, and otherwise its language and activity,
such as Japanese reading. Its lead names the language and the activity; on
another learner's log with a known display name, it also names who logged it,
linked to their profile. It shows the measure, rounded and pluralized as the
lists show it, the score, and the UTC day and time, both read from the creation
instant. Labelled rows then show the description, unless it is empty or already
the heading, and the tags, unless there are none; the record's ID is not shown. It offers its owner
and administrators Edit log and Delete log; other visitors see neither.

A log tagged `nsfw`, in any case, never shows its description as a title or a
link label: the detail page is titled by its language and activity, and the
edit and contest pages link back by them. Wherever the description itself is
shown, on the detail page, the contest page's summary and the activity feeds,
`frontend/apps/webv3/src/components/LogDescription.tsx` masks it with Paper's
`Spoiler`, as Webv2 does: the text is blurred under an NSFW label and hidden
from assistive technology until the reader selects Reveal NSFW description, and
a mouse resting on it previews it. It stays masked until the page hydrates.
`frontend/apps/webv3/src/lib/format.ts` owns the tag rule and the name links
use.

`/logs/$logId/edit` shows the edit form only to the log's owner and
administrators; another account
sees that the record is someone else's, and an anonymous visitor is asked to
sign in. It uses the same tags and tag sidebar, shows the activity read-only
and keeps contest submissions as they are; Cancel and Save return to the log.
A log submitted to no contest, ongoing or ended, offers its language; once
submitted, the language is read-only because the API refuses to change it. The
language, measure,
time spent, description and tags change through the update operation, which
rescores the log's submissions to unfinished contests and leaves finished
contests' scores unchanged. Timed logs saved with their minutes as an amount are
edited and saved as a duration. A save replaces the edit page with the updated
detail. Retrying an unconfirmed update is safe. The edit draft starts from the
log as first read; later reads do not replace it.

`/logs/$logId/contests` lets the log's owner, or an administrator, choose which
of the owner's ongoing contest registrations the saved log counts in; another
account sees that the record is someone else's, and an anonymous visitor is
asked to sign in. The API accepts only the owner's ongoing registrations, so
`ownerRegistrationsQuery` in `frontend/apps/webv3/src/lib/logs.ts` reads them for
whoever is viewing: the owner's own ongoing registrations, or for an
administrator the owner's contest registrations of this year and last whose
contest has started and has not finished by today's UTC date, the API's
ongoing rule. Those registrations do not name a community contest's organizer,
so their cards omit it. A choice the API no longer counts as ongoing is refused
as invalid. Each
registration is offered unless the contest does not allow the log's activity,
the registration lacks its language or the contest has ended, in that order.
As in Webv2, the page opens with a summary of the log and lists each
registration as a separate card naming its organizer and the points the log adds;
offered contests come first, and an unavailable one shows its reason on its
card. The owner's current submissions start
selected; with `?preselect=1`, a log not yet submitted anywhere starts with every
eligible registration selected, as Webv2's link after saving does. Saving sends
the selection to the contest registration update operation, which adds and removes submissions to unfinished contests and keeps
submissions to finished contests and their saved scores; the page lists those
as finished. A save opens the updated detail, and saving again is safe. Without
an eligible registration the page links to the contest lists. Removing a
submission leaves the personal record unchanged.

Delete log asks for confirmation, then deletes the log and opens the owner's
[profile statistics](webv3-profile.md). A log already gone counts as deleted. The API refuses to
delete a log submitted to a finished contest.

Invalid input focuses the first invalid control. Validation follows the API's
text rules: a description allows 255 characters, and tags are normalized as Go's
`strings.TrimSpace` and `strings.ToLower` do and deduplicated before the limits
of 10 tags and 50 UTF-8 bytes per tag apply. Unicode white space is trimmed and
each character is lowercased on its own, so `İ` becomes `i` and `Σ` becomes `σ`
in every position. A pending save disables
controls and repeated submits. Rejected saves keep the draft. Each save, update and deletion checks
the current session against the account that started the draft or viewed the log; an account
switch is refused before the API write, updates the header and offers an
explicit new draft for the new account. The first signed-in account anchors the
draft even when the page started anonymously and login happened in another tab.
An unavailable account check reports that nothing was saved. An expired session
offers login in a new tab so the original form survives. An ambiguous transport
failure asks the user to check their activity before retrying; the API has no
idempotency key, so retrying an ambiguous save can create another record. Drafts
are held in memory and are lost when the form is reloaded or closed.

`logging.spec.ts` covers the form, saves, the contest step after saving with
and without ongoing registrations, detail, ownership and tenant isolation;
`logging-form.spec.ts` covers tag chips and suggestions, the tag sidebar and its
limit, remembered choices, unavailable storage and the edit form's tags; `logging-edit.spec.ts` covers edits, deletion, owner actions and the
score estimate; `logging-contests.spec.ts` covers submitting a saved log to a
contest, removing it and refusing other visitors; `logging-nsfw.spec.ts` covers
NSFW masking on the detail page, its back links and the activity feed;
`finished-contest.spec.ts` covers an edit, a deletion and the contest step
for a log in the seeded finished round, and a contest's last day;
`logging-admin.spec.ts` covers an administrator editing, resubmitting and
deleting another learner's log, and refusing that edit after an account switch;
`logging-session.spec.ts` covers expiry, CSRF, suspension and
account switches. Run them with `WEBV3_BASE_URL`, `TADOKU_API_URL`, `E2E_READER_EMAIL`
and `E2E_READER_PASSWORD` for the documented development fixture. Ownership,
suspension and account-switch cases also use `E2E_SECOND_EMAIL` and
`E2E_SECOND_PASSWORD` for Dev Admin; the suspension case first checks that the
administrator reaches the selected branch and restores the Reader role there.
Journeys get their accounts from the shared fixtures, which keep both
accounts' sessions across runs and sign in through authv2 with a newly
registered learner; [Credentialed runs and fixture accounts](webv3-testing.md#credentialed-runs-and-fixture-accounts)
describes them and the login budget. Set `NODE_EXTRA_CA_CERTS` to the trusted Lab CA
and `PLAYWRIGHT_OUTPUT_DIR` outside the checkout. Authenticated runs disable
traces and videos; browser-boundary delay and abort cases cover repeated submits
and transport errors without mocking the API's successful writes.
