# Tadoku feature map

Tadoku lets language learners log activities, track progress and participate in
contests and leaderboards. The main site is `webv2`; `auth` handles accounts and
`admin` manages content and moderation. They use the native `tadoku-api` backend.

Start with [the verification skill](../../SKILL.md) for environment selection and
safety. Read only the area your task affects. These are **development** URLs and
user journeys, not deployment metadata or an assertion that every check has passed.
Generate selected links with `dev url`; bare links below use your browser's current
cookie (base in a fresh context). Never use production `tadoku.app` for these checks.

| User goal | Entry point | Read |
| --- | --- | --- |
| Sign in, register, recover access, change account settings | Main-site Log in/Sign up; account site | [Accounts](accounts.md) |
| Record learning, edit logs, inspect a learner's progress | Log activity; Profile → Statistics/Updates | [Activity and profiles](activity.md) |
| Join/create contests, compare scores and rankings | Contests; Leaderboard | [Contests and leaderboards](contests.md) |
| Read the homepage, blog, manual and announcements | Home; Blog; footer links | [Public content](content.md) |
| Publish content, manage languages, moderate users and feature access | User menu → Admin | [Administration](admin.md) |

## Fixtures and test boundaries

Hosts: main `https://tadoku.dev.lab`, accounts `https://account.tadoku.dev.lab`,
admin `https://admin.tadoku.dev.lab`. See [configuration](../../../../../.dev/config.yaml).
The shared synthetic accounts are `reader@tadoku.app` (Dev Reader) and
`dev@tadoku.app` (Dev Admin), with development fixture password `tadoku` unless
overridden outside Git. These email strings are **not** test destination hosts.
Don't change their credentials/profile/roles for an unrelated test.

Branch `migrate`/`seed` tasks reuse those identities and populate branch-local data.
Missing shared identities require authorized setup from the [runbook](../../../../../.dev/README.md),
not invented IDs or auth bypasses. User IDs are generated; obtain them from the
profile link or your authenticated session, not a hardcoded UUID.

Useful fixtures from [the seed sources](../../../../../scripts/dev/seed/):

- Official contest `00000000-0000-4000-8000-000000000101`, “Dev Tadoku Round”.
- Private contest `00000000-0000-4000-8000-000000000102`, owned by Dev Reader.
- Public page `/pages/dev-welcome`, post `/blog/posts/dev-round-open`.
- Reader/admin logs with reading/listening examples; discover them through profiles.

Seeds use current dates and reruns update records. Don't assert fixed scores,
timestamps, registration windows or ordering against a shared/previously used DB.
Record starting values; create uniquely named records in your selected branch when
testing writes. Re-seeding is not a reset. Account providers remain shared.

## Sweep and maintenance

For broad application changes, cover anonymous home/content → real login/return →
contest leaderboard → owned log write/read/profile/leaderboard → admin authorization
and affected content → logout. For isolated edits, cover just relevant journeys and
adjacent permission/persistence states. Repeat with separate selected/base contexts
when routing or isolation changes. Never mutate shared fixtures for a broad sweep.

Token-reflector is base-only infrastructure, not a separate user journey. Paper
playground/styleguide are design tooling, not the deployed product; live overlays
are deferred. No check here implies those previews have been verified.

Maintain the relevant entry alongside behavior changes: entry point, prerequisites,
observable outcome, failure/empty/permission states, traps and source anchors. Keep
source pointers secondary; Bazel metadata remains the only deployable graph.
This behavior-first structure is inspired by
[poteto's verification-skill example](https://github.com/poteto/verification-skill-example).
