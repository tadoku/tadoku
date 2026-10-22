# Administration

## Find it

Sign in as a development administrator, then user menu → **Admin**. The admin
host's sidebar exposes Posts, Pages, Announcements, Languages and Users.
Content routes use namespace `tadoku`: `/posts/tadoku`, `/pages/tadoku`,
`/announcements/tadoku`, each with `/new` and record edit routes.
Platform/moderation routes are `/languages` and `/users`. User row action menus
contain Feature access and, for non-admin users, Ban/Unban User.

## Verify

- First select **both admin and main host links** in the same context and prove
  the API response's `X-Dev-Selected` names your route before any write. A
  frontend-only branch uses the base API with its branch tenant; require an API
  overlay only for changed API code. Loading the admin overlay alone does not
  select the main hostname. Check real administrator access; a reader
  and anonymous session must not gain management rights through a visible button.
- Content: create an owned draft with a unique slug/title, preview, edit, save
  and reload. Exercise publish/unpublish/delete only when in task scope, and
  confirm the corresponding [public page/post](content.md) in that branch.
  Pages use HTML content; posts use the editor's Markdown path. Exercise empty
  fields, cancel and publication state rather than only a success toast.
- Announcements: verify title/body/style/link and start/end window, reload the
  edit form, then observe the matching public display. Use owned records only.
- Languages: filter by name/code, check no-match and clear, open Add/Edit Language
  and cancel. For authorized write tests, use an owned branch record; verify
  validation and persisted display name. Existing language codes aren't editable.
- Users: search by name/email, clear and verify empty results. For moderation or
  feature-access changes, use a disposable identity, confirm/cancel the action,
  reload, and verify the affected user's real journey in another context. Record
  both UI state and API result. Do not ban shared fixture users or alter their roles.

## Traps

Content namespace and environment selection are different. Changing namespace
doesn't switch branches. Accounts, application records, roles and
named-user Flipt access use the selected
branch tenant and its `test/tadoku_<route>` namespace. Dev Admin is an
administrator there through its own branch grant; administrators and bans in
`tadoku/prod` do not apply. Do not
change shared credentials or canonical roles through a branch test. Feature
access is distinct from global Flipt flag policy. Don't modify
global flags or expose its administration interface to make a UI test pass.
CMS copy-only requests are not application-code tasks; obey `AGENTS.md` rather
than checking those edits into Git. Record retained test records instead of
deleting data you did not create.

## Source anchors

[Admin routes](../../../../../frontend/apps/admin/pages/),
[sidebar](../../../../../frontend/apps/admin/app/ui/DashboardLayout.tsx),
[content editors](../../../../../frontend/apps/admin/app/content/),
[feature access](../../../../../frontend/apps/admin/app/feature-access/).

## Retained email API

Administrators can inspect their selected tenant with `GET /api/internal/emails`
and `GET /api/internal/emails/{id}` on its branch host once its email gateway
rule is installed. An unconfigured gateway returns 404 before the API; retain
that result as an unmet public-route prerequisite. Attached `tadoku-tenant
access-token create` can mint a branch `email:read` token without a shared login.
Confirm token list/detail reads, rejection on another business route and another
scoped API branch, and expiry/revocation. Retain only sanitized status/body
assertions and metadata; never retain the token or pass it through Job logs.
Cookie-admin verification needs an already approved session or owned provider
fixtures; disposable HTTP gateway E2Es complement the required live branch gate. There is no email UI or
public send endpoint. Verify a produced email becomes `held` in development,
read back its template/params as Dev Admin, and confirm Dev Reader and anonymous
requests are denied. Check `X-Dev-Selected` and the actual API overlay backend.
Production rows require an audit entry for each returned message; use disposable
HTTP E2Es for production-tenant audit and SMTP delivery cases, never production
email or shared account mutations. See the [email feature guide](../../../../../docs/docs/tadoku-api/email.md).
