---
title: Webv3 headers, indexing and sharing
description: The security headers and report-only CSP Webv3 sends, how its pages name their canonical web-host URL and keep other hosts out of search results, and the metadata search results and link previews read.
sidebar_position: 3
---

# Webv3 headers, indexing and sharing

Read this when you add an inline script, a page or a host to Webv3, or change
what search engines and link previews see.

## Security headers

`frontend/apps/webv3/src/start.ts` gives every TanStack Start response, pages
and server functions:

| Header | Value |
| --- | --- |
| `Strict-Transport-Security` | `max-age=31536000`, without `includeSubDomains` or `preload`, so it binds only the host that sent it |
| `X-Content-Type-Options` | `nosniff` |
| `Referrer-Policy` | `strict-origin-when-cross-origin` |
| `Content-Security-Policy` | `frame-ancestors 'none'`, enforced |
| `Content-Security-Policy-Report-Only` | The full policy, with the request's script nonce |

The full policy is report-only until its violations have been read. It names
no report endpoint, so a browser reports violations only in its console. Built
assets and the Nitro routes in `frontend/apps/webv3/server/routes/`, such as
`/healthz` and `/robots.txt`, do not get these headers.

Every response gets a new nonce. `frontend/apps/webv3/src/router.tsx` passes
it to the router, which writes it on its own scripts and in a `csp-nonce` meta
tag that the browser's router reads back. An inline script must get the nonce
the same way: render it with the router's `ScriptOnce`, as the theme bootstrap
in `frontend/apps/webv3/src/routes/__root.tsx` does, never as a plain
`<script>`.

## Indexing

A page belongs on its web host: the deployment's `TADOKU_WEB_URL` host, or the
matching branch web host, as `isWebHost` in
`frontend/apps/webv3/src/lib/urls.server.ts` decides. Every page links its
canonical URL, the web host with the page's path and without its query or
fragment, and gives the same URL as `og:url`.

| Request host | `X-Robots-Tag` | `/robots.txt` |
| --- | --- | --- |
| The web host | None | `Disallow:` with no path, so crawlers may read everything |
| Any other, such as `next.` or a branch host | `noindex` | `Disallow: /` |

While Webv3 serves `next.tadoku.app`, every production page is therefore
noindex and canonical to `tadoku.app`. Once `tadoku.app` routes to Webv3 its
pages are indexable, provided the route that serves it adds no `X-Robots-Tag`
of its own.

## Sharing metadata

Pages a visitor would share, CMS pages, blog posts, contests, contest profiles
and profiles, set their head with `sharingMeta` from
`frontend/apps/webv3/src/lib/page-meta.ts`: the title, a description of at
most 160 characters and the matching Open Graph and Twitter card tags. A
description comes from the page's own content where it has some, such as a CMS
page's text, a post's Markdown or a contest's description.

Private and action pages, such as forms, a learner's own lists, logs and log
out, add `noindex` from the same module, so they stay out of search results on
the web host too. So do a private contest's standings, activity and participant
profiles, which anyone with the URL can open. `sharingMeta` cuts every
description to 160 characters.

## Checks

`frontend/apps/webv3/e2e/indexing.spec.ts` checks the headers, the nonce, the
canonical link, robots and the metadata. Its web-host case requests the server
with the web host's `Host` header, so it runs only against a plain-HTTP server
such as the image smoke and skips on a branch host.
