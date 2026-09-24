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
  the API response reaches your branch before any write. Loading the admin
  overlay does not isolate API data. Check real administrator access; a reader
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
doesn't switch branches. Account/permission/feature providers may be shared even
when application records use a branch DB; don't assume all admin actions are
isolated. Feature access is distinct from global Flipt flag policy. Don't modify
global flags or expose its administration interface to make a UI test pass.
CMS copy-only requests are not application-code tasks; obey `AGENTS.md` rather
than checking those edits into Git. Record retained test records instead of
deleting data you did not create.

## Source anchors

[Admin routes](../../../../../frontend/apps/admin/pages/),
[sidebar](../../../../../frontend/apps/admin/app/ui/DashboardLayout.tsx),
[content editors](../../../../../frontend/apps/admin/app/content/),
[feature access](../../../../../frontend/apps/admin/app/feature-access/).
