---
sidebar_position: 2
title: Content API
---

# Content API

The Content API is the public namespace for Tadoku's blog, page, and announcement
operations. These routes are served by Tadoku API.

- Fetch, create, update, and delete blog posts
- Fetch, create, update, and delete web pages
- Fetch, create, update, and delete announcements

All content uses markdown.

## Contract and persistence

- The canonical OpenAPI contract is `services/tadoku-api/spec/openapi.yaml`.
- Content is stored in PostgreSQL through Tadoku API's page, post, and announcement features.
- Content schema history is part of the consolidated Tadoku API migration set.

## Important links

- [Source code](https://github.com/tadoku/tadoku/tree/main/services/tadoku-api)
- [API reference](../api/content/content-api)
- [OpenAPI source](https://github.com/tadoku/tadoku/blob/main/services/tadoku-api/spec/openapi.yaml)
