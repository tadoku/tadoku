---
title: Webv3 contest creation and scoring
description: Webv3's contest creation form and administrator scoring page, their viewer gates, validation, writes, failure states and browser journeys.
sidebar_position: 3
---

# Webv3 contest creation and scoring

Read this when you change Webv3's contest creation form or contest scoring page.

## Contest creation

`/contests/new` asks anonymous visitors to sign in. A signed-in account sees the
form when the API's create-permission check answers 200; when it answers 403 the
page explains that accounts can create contests from one month after joining, up
to 12 a year. The contest lists show their Create contest action on the same
answer, as in Webv2.

The form takes a title, an optional description, start, end and sign-up dates as
UTC calendar days, allowed languages (none allows every language) and allowed
activities, starting with the API's default activities. Accounts the API lets
create official rounds can mark the contest official; community contests can
instead be open only to those with the link. Official rounds stay public and
allow every language.

- `frontend/apps/webv3/src/lib/contest-form.ts` owns the form's values, their
  validation with the API's rules and the `NewContest` the create server function
  accepts, whose official and community variants keep those combinations apart.
  A title needs more than three and at most 255 characters, a description at most
  500, the end date must not precede the start date, and contests by
  non-administrators must not start or end before the current UTC day.
- `frontend/apps/webv3/src/lib/contests.ts` owns the editor query, which reads the
  permission and the configuration options together, and `useCreateContest()`.
  `frontend/apps/webv3/src/lib/contests.server.ts` calls them and the create
  operation through the generated client.
- `frontend/apps/webv3/src/components/ContestForm.tsx` composes the form from
  `usePaperForm` and Paper's bound controls, and owns navigation and failure
  messages.

Invalid input focuses the first invalid control. As with
[activity logging](webv3-logging.md), the first signed-in account owns the draft,
and the server refuses to create the contest when the session now belongs to
another account. A created contest replaces the form with its
[contest page](webv3-contest.md). An ambiguous failure asks the user to check My contests before
trying again, because the API has no idempotency key.

## Contest scoring

`/contests/$id/scoring` follows Webv2's page: anonymous visitors are redirected to
log in and returned afterwards, and every account except an administrator sees
that the feature is not yet available. The API also lets a contest's organizer
change its scoring; the page does not offer that. The server functions refuse
scoring writes unless the session still belongs to the administrator who opened
the page.

The page reads the contest, the logging configuration and the platform and
contest rule-set versions. Until the contest's first UTC day, an administrator
can draft a version that overrides platform rules with a pinned published
platform version or replaces them. A draft holds ordered rules that match an
activity and optionally a unit, language and tag, and score the amount or the
minutes at a rate; stackable modifiers multiply the matching base rule. Rules are
sent with priorities in their order. Each version lists its rules and status;
a draft can be published, which makes it immutable, and a published version can
be activated for new contest submissions, each after confirmation. From the
contest's first UTC day the page shows that scoring is locked, and the API refuses
changes.

- `frontend/apps/webv3/src/lib/scoring-draft.ts` owns the draft's values,
  validation with the API's rate and tag rules and the `NewRuleSet` whose
  override and replace variants carry or omit the pinned version.
- `frontend/apps/webv3/src/lib/scoring.ts` owns the server functions, the page
  query and `useChangeScoring()`, which reads the versions again after a change.
  `frontend/apps/webv3/src/lib/scoring.server.ts` calls the operations through the
  generated client.
- `frontend/apps/webv3/src/components/ContestScoring.tsx` and
  `frontend/apps/webv3/src/components/RuleSetForm.tsx` compose the page.

## Checks

`frontend/apps/webv3/e2e/contest-admin.spec.ts` checks the anonymous prompt and
scoring redirect, then, with both fixture accounts, creates a contest through the
form after each invalid field is focused and compares it with the API, shows Dev
Reader the refusal after using up its yearly allowance in the selected tenant,
drafts, publishes and activates a rule set and compares it with the API, checks a
started contest's lock, and that Dev Reader cannot reach scoring, even for its
own contest, and that only administrators see the contest page's Scoring rules
link. Pages are checked at 390px. The anonymous checks also run in the
production container smoke.
