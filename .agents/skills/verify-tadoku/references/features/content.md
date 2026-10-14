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

## Webv3 preview

Generate its selected link with `dev url --owner <owner> --host next.tadoku.dev.lab '/'`.
The homepage reads its branch's official contest and standings through the matching
Webv2 API gateway; `/plan` saves only a local browser goal. Check product anchors in
the server HTML and after hydration, then follow Daily goal and verify the links
still use `<route>.tadoku.dev.lab`. Log activity must reach that same Webv2 branch,
including its account login redirect. Compare two branch hosts in one browser
profile to detect selection leaks; do not change shared identities or credentials.

Check the unavailable-data state and `/healthz` when its configured API fails.
The container smoke suite proves runtime URL overrides and concurrent Host isolation
with synthetic read-only data; live ingress, tenant selection and login require
separate development verification. Production canonical `next.tadoku.app` uses
`tadoku.app` links, while `<route>.next.preview.tadoku.app` uses
`<route>.preview.tadoku.app` links.

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
