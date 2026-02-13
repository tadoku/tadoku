---
sidebar_position: 3
title: webv2
---

# Webv2 frontend

This is the main frontend of the Tadoku website. It contains all logging & contest functionality.

## Architecture

- A [Next.js](https://nextjs.org/) app written using TypeScript
- Data is fetched through [React Query](https://tanstack.com/query/latest/docs/framework/react/overview) and response is validated with [Zod](https://zod.dev/)
- Consumes the following APIs:
  - immersion-api: the main api for logging and contest functionality (specs: `services/immersion-api/http/rest/openapi/api.yaml`)
  - content-api: the api for fetching blog posts & pages (specs: `services/contest-api/http/rest/openapi/api.yaml`)
- The app uses Tailwind CSS for styling
- Uses the Tadoku component library "ui" from within the workspace
  - Refer to `frontend/apps/styleguide` for a reference on how to use this component library
- All forms are written with [react-hook-form](https://react-hook-form.com/)
- Dates are manipulated using [Luxon](https://github.com/moment/luxon/)

## Important links

- [Source code](https://github.com/tadoku/tadoku/tree/main/frontend/apps/styleguide)
- [Production](https://tadoku.app/)
