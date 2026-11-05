---
title: Webv3 activity logging
description: Native Webv3 activity creation, saved detail, edits, deletion and the score estimate, their draft owners, session safety and browser journeys.
sidebar_position: 3
---

# Webv3 activity logging

Read this when you change a Webv3 log page, the log form or its data.

`/logs/new` prompts anonymous visitors to sign in through authv2, then loads
configuration and ongoing contest registrations from the selected tenant.
Language and activity determine the available unit or duration input and
eligible contests. Amount activities also accept an optional time spent. Eligible contests are selected initially; a user can uncheck
them to save a personal record. Changing the activity clears the quantity,
because activities count in different units. A missing unit is reported on the
amount field, which shows one error for the amount and its unit.

The draft has three owners:

- `frontend/apps/webv3/src/lib/log-draft.ts` owns the form's values: their
  defaults for a new or saved log, the unit and contests that follow a language
  or activity change, validation with the API's rules, and conversion to the
  `NewLog`, `LogUpdate` and score preview inputs the server functions accept.
- `frontend/apps/webv3/src/components/LogForm.tsx` composes the new and edit
  screens from `usePaperForm` and Paper's bound controls, and owns navigation,
  the score estimate and failure messages.
- `frontend/apps/webv3/src/lib/logs.ts` owns the server functions, their query
  keys and `useSaveLog()`, `useUpdateLog()` and `useDeleteLog()`, which cache a
  saved log for its detail page and refresh the viewer after an account switch.

`frontend/apps/webv3/src/components/DraftOwnerGate.tsx` fixes the account that
owns a page's draft, for both forms.

`frontend/apps/webv3/src/lib/logs.server.ts` calls the configuration,
registrations, create, detail, update, delete and score preview operations
through the generated client, which validates their responses, with the session
cookies described in [account sessions](webv3.md#account-sessions). A registration without its ID or contest makes the
options unavailable. The API owns scoring, contest eligibility, authorization
and persistence; Webv3 never calculates a score itself.

Both forms show the API's estimated score once the draft is valid and has not
changed for 300 ms. The request is a GET server function and sends no contest
registrations, so it estimates the personal score only. An incomplete draft, a
pending first estimate and a failed one show a dash; a zero estimate shows N/A.
A saved amount names its unit by the configuration's `unit_key`; the API resolves
the language's variant.

A successful save navigates to `/logs/$logId` using the returned ID and displays
the persisted language, measure, tags, score and contests. Reload reads that
record again. An expired session cookie retries a public detail read
anonymously. Public detail omits private contest submissions; only the owner
gets the contest management link to Webv2, which keeps contest submissions and
history.

The detail page offers its owner Edit log and Delete log; visitors see neither.
`/logs/$logId/edit` shows the edit form only to the log's owner; another account
sees that the record is someone else's, and an anonymous visitor is asked to
sign in. Language, activity and contest submissions stay fixed. The measure,
time spent, description and tags change through the update operation, which
rescores the log's submissions to unfinished contests and leaves finished
contests' scores unchanged. Timed logs saved with their minutes as an amount are
edited and saved as a duration. A save replaces the edit page with the updated
detail. Retrying an unconfirmed update is safe. The edit draft starts from the
log as first read; later reads do not replace it.

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

`logging.spec.ts` covers the form, saves, detail, ownership and tenant
isolation; `logging-edit.spec.ts` covers edits, deletion, owner actions and the
score estimate; `logging-session.spec.ts` covers expiry, CSRF, suspension and
account switches. Run them with `WEBV3_BASE_URL`, `TADOKU_API_URL`, `E2E_READER_EMAIL`
and `E2E_READER_PASSWORD` for the documented development fixture. Ownership,
suspension and account-switch cases also use `E2E_SECOND_EMAIL` and
`E2E_SECOND_PASSWORD` for Dev Admin; the suspension case first checks that the
administrator reaches the selected branch and restores the Reader role there.
The identity API admits ten logins per account in 15 minutes, counting
successful ones. `frontend/apps/webv3/e2e/logging-fixtures.ts` therefore shares
one real session per account, held in memory, across journeys where login is
incidental. The logging files and the account E2E together sign in with Dev
Reader eight times and Dev Admin five times, so wait 15 minutes after the last
login before repeating a complete run. A failed test restarts the Playwright
worker, whose shared sessions then sign in again, and rejected attempts count
towards the limit too. Set `NODE_EXTRA_CA_CERTS` to the trusted Lab CA
and `PLAYWRIGHT_OUTPUT_DIR` outside the checkout. Authenticated runs disable
traces and videos; browser-boundary delay and abort cases cover repeated submits
and transport errors without mocking the API's successful writes.
