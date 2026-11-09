---
title: Webv3 content pages
description: Webv3's manual, CMS pages and blog, their paths, how CMS HTML and post Markdown are sanitized, not-found and unpublished states, and their browser journey.
sidebar_position: 3
---

# Webv3 content pages

Read this when you change Webv3's manual, a CMS page or the blog.

The paths match Webv2's. `/pages/manual` is the manual, `/pages/$slug` is any
other CMS page, such as `about`, `faq`, `contact` or `privacy`, `/blog` lists
posts with an optional trailing page number such as `/blog/2`, and
`/blog/posts/$slug` is one post. The navigation's Guide link opens the manual,
and the footer links About, the manual, the rules, questions (`faq`), privacy and
contact on the same host, in Webv2's order; the blog is in the main navigation.

- `frontend/apps/webv3/src/components/ManualPage.tsx` holds the manual. Its
  text is application code ported from Webv2; change it there.
- CMS page bodies are edited in the admin CMS only. Webv3 renders whatever the
  API returns for the slug and holds no page copy.
- `frontend/apps/webv3/src/lib/content.ts` owns the server functions and query
  options. `frontend/apps/webv3/src/lib/content.server.ts` reads pages, posts and
  post lists anonymously through the generated client.
- `frontend/apps/webv3/src/components/ContentPages.tsx` composes the CMS page,
  the blog list and a post.

## Rendering rules

Webv3 renders CMS HTML only after `content.server.ts` has passed it through
DOMPurify's default sanitization, the same call Webv2 makes in the browser. The
server runs DOMPurify on a jsdom window. Nitro bundles server code, but jsdom
reads files beside its sources, so `frontend/apps/webv3/vite.config.ts` traces
jsdom into the build output as a dependency instead. The `SanitizedHtml` type
marks the only string the CMS page renders as HTML.

Post bodies are Markdown, rendered as Webv2 renders them: `react-markdown` with
GitHub-flavored Markdown, without raw HTML, through `rehype-sanitize`'s default
schema. The blog's first page leads with the newest post's plain-text excerpt
of up to 500 characters, cut at a word boundary as Webv2 does. Publish dates
show the UTC day.

## States

The API answers 404 for a missing or unpublished page or post, and refuses an
ID-shaped slug, which it reserves for administrators, to anyone else. Webv3
shows each of these as not found with HTTP status 404. A post without a publish
time, or any other unreadable answer, shows an unavailable state with a retry.
The blog holds 10 posts per page. A missing or unreadable page number is page 1,
as on the leaderboards, and a page past the last says so and links to the newest
posts. An empty blog says no posts exist yet.

## Checks

`frontend/apps/webv3/e2e/content.spec.ts` checks the server-rendered manual and
its navigation without JavaScript, compares the seeded `dev-welcome` page's text
with the API's HTML and checks that nothing DOMPurify removes reaches the page,
checks 404s for unknown slugs, compares blog pages with the API, follows the
pagination and opens a post, and checks containment at 390px. With Dev Admin's
credentials it also creates content in the selected tenant: enough published
posts for two pages, a draft post and page that must stay not found, and a page
with a script, an event handler and a `javascript:` link that must not survive.
These records stay in the branch tenant, and later runs reuse them.
