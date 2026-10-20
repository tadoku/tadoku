# Tadoku feature map

Tadoku lets language learners log activities, track progress and participate in
contests and leaderboards. The main site is `webv2`; `auth` handles accounts and
`admin` manages content and moderation. They use the Tadoku API backend.

Start with [the verification skill](../../SKILL.md) for environment selection and
safety; [Development environment](../../../../../docs/docs/develop/environment.md)
covers hosts and branch selection, and
[Development data](../../../../../docs/docs/develop/development-data.md#seed-data)
covers fixture accounts. Read only the area your task
affects. These are **development** URLs and user journeys, not deployment metadata
or an assertion that every check has passed. Generate selected links with `dev url`;
bare links below target the base; branch links carry the route in their hostname.
Never use production `tadoku.app` for these checks.

| User goal | Entry point | Read |
| --- | --- | --- |
| Sign in, register, recover access, change account settings | Main-site Log in/Sign up; account site | [Accounts](accounts.md) |
| Record learning, edit logs, inspect a learner's progress | Log activity; Profile → Statistics/Updates | [Activity and profiles](activity.md) |
| Join/create contests, compare scores and rankings | Contests; Leaderboard | [Contests and leaderboards](contests.md) |
| Read the homepage, blog, manual and announcements | Home; Blog; footer links | [Public content](content.md) |
| Publish content, manage languages, moderate users and feature access | User menu → Admin | [Administration](admin.md) |

## Fixtures and test boundaries

“Main”, “account” and “admin” host below mean `tadoku.dev.lab`,
`account.tadoku.dev.lab` and `admin.tadoku.dev.lab`. Use the shared synthetic
accounts Dev Reader and Dev Admin from the
[fixture accounts](../../../../../docs/docs/develop/development-data.md#seed-data). Their
`@tadoku.app` email strings are **not** test destination hosts. Don't change their
credentials/profile/roles for an unrelated test.

The automatic `tenant` hook creates native v2 accounts using those fixture UUIDs and populates branch-local
data in `tadoku/<route>`. Frontend-only branches use the base API with that
selected tenant. Use branch-host links for auth or admin API writes and
check the API response's `X-Dev-Selected`.
Missing base fixtures require authorized setup. Branch accounts keep the base
fixture UUIDs, names and documented password, with independent credentials and
sessions. Obtain actual IDs from the session or profile, not a hardcoded UUID.
Provision seeds `identity-v2=true` only in the owned branch namespace.

Useful fixtures from [the seed sources](../../../../../scripts/dev/seed/):

- Official contest “Dev Tadoku Round”; discover its ID through the selected
  branch's contest list or UI. Only canonical `tadoku/prod` uses
  `00000000-0000-4000-8000-000000000101`.
- Private contest owned by Dev Reader; discover its ID in that branch. Only
  canonical `tadoku/prod` uses `00000000-0000-4000-8000-000000000102`.
- Public page `/pages/dev-welcome`, post `/blog/posts/dev-round-open`.
- Reader/admin logs with reading/listening examples; discover them through profiles.

Seeds use current dates and reruns update records. Don't assert fixed scores,
timestamps, registration windows or ordering against a shared/previously used DB.
Branch contest, registration and log UUIDs are derived from the full tenant key
with UUIDv5. A canonical fixed UUID will return 404 on a branch; obtain the
branch's own IDs before opening a deep link.
Record starting values; create uniquely named records in your selected branch when
testing writes. Re-seeding is not a reset. Native accounts are tenant-owned; v1 providers remain shared.

## Sweep and maintenance

For broad application changes, cover anonymous home/content → real login/return →
contest leaderboard → owned log write/read/profile/leaderboard → admin authorization
and affected content → logout. For isolated edits, cover just relevant journeys and
adjacent permission/persistence states. Repeat with separate selected/base contexts
when routing or isolation changes. Never mutate shared fixtures for a broad sweep.

Token-reflector is base-only infrastructure, not a separate user journey. Paper
playground/styleguide are design tooling, not the deployed product; they have no
live overlays. No check here implies those previews have been verified.

Maintain the relevant entry alongside behavior changes: entry point, prerequisites,
observable outcome, failure/empty/permission states, traps and source anchors. Keep
source pointers secondary; Bazel metadata remains the only deployable graph.
This behavior-first structure is inspired by
[poteto's verification-skill example](https://github.com/poteto/verification-skill-example).
