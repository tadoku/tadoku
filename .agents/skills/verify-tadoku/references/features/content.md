# Public content and navigation

## Find it

Main-host `/` is the homepage. **Blog** opens `/blog`, with posts at
`/blog/posts/<slug>`. Footer/content links lead to `/pages/<slug>`; `/pages/manual`
is a code-owned manual route rather than the generic CMS page. Seeded examples are
`/pages/dev-welcome` and `/blog/posts/dev-round-open`. Announcements can link into
those pages while their display window is active.

## Verify

- In a fresh anonymous context, open the homepage and follow the affected
  navigation link. Check loaded content and relevant API responses, not just the
  page shell. Exercise desktop and narrow viewport navigation when layout changes.
- Read a blog post and CMS page, reload and return through navigation. Check
  titles, body formatting and links. Test missing slugs/empty lists in an owned
  branch fixture when affected; an API error must not look like successful content.
- For announcement changes, verify visibility, style and destination for the
  relevant active/inactive window using an owned fixture; absence alone does not
  prove a renderer bug. For publication changes, follow the [admin journey](admin.md)
  and check the public result in the same selected API environment.
- For HMR, leave the target page open and edit application UI code; confirm the
  visible update without navigation and with the same Pod UID/image. Don't use
  CMS copy edits as a source-sync demonstration.

## Traps

About, Contact and similar generic page bodies are CMS-owned. For copy-only
requests, follow `AGENTS.md`: stop coding and report that the edit belongs in the
admin CMS. Never patch frontend text, SQL fixtures or migrations to change that
copy. Missing FAQ/About/Contact seed content is not proof the branch failed;
start with the known seeded page. Home combines backend content and leaderboards;
API-only overlays must work with the base frontend, including server rendering.
Reject accidental requests to production hosts during development verification.

## Source anchors

[Home](../../../../../frontend/apps/webv2/pages/index.tsx),
[blog](../../../../../frontend/apps/webv2/pages/blog/),
[pages](../../../../../frontend/apps/webv2/pages/pages/),
[navigation](../../../../../frontend/apps/webv2/app/ui/Navigation.tsx),
[content seeds](../../../../../scripts/dev/seed/content.sql).
