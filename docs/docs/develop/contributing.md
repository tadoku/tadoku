---
title: Contributing workflow
description: Who may contribute to Tadoku, how to shape commits and pull requests, and the rules for migrations, CMS content, bug reports and evidence.
sidebar_position: 4
---

# Contributing workflow

Read this when you are about to commit, open a pull request or pick up a bug report.

## Outside contributions

Tadoku does not accept pull requests from outside contributors; unsolicited
pull requests are generally closed without merge. Report bugs and ideas as a
GitHub issue or on the Tadoku Discord server, and report security issues
privately to the maintainer. `CONTRIBUTING.md` at the repository root has the
details. The rest of this page applies to maintainers and invited contributors.

## Commits

- Make each commit one logical change. Do not bundle unrelated changes.
- Split larger refactors that span many files into coherent chunks, such as one
  commit per page, per service or per domain area.

## Pull requests

- Run the checks and gather the evidence described in
  [Verifying changes](./verifying-changes.md). Attach evidence to the pull
  request.
- Explain reviewed HTTP golden changes and any intentional import-boundary
  change in the pull request body.

## Migrations ship alone

A schema or data migration lands on `main` in its own commit and pull request
and is deployed before any code that depends on it. CI rejects pull requests
that mix migration files with other changes. See
[Tadoku API database](../tadoku-api/database.md).

## CMS-managed content

Public page copy, such as the Contact and About pages, lives in the CMS. Edit it
in the admin CMS only. Never ship that copy as a frontend change, a migration or
any other SQL rewrite. When a request turns out to be page copy rather than
application code, stop and report that the edit belongs in the admin CMS.

## Bug reports

1. Reproduce the bug first with the smallest useful existing test, build, lint
   check or manual scenario.
2. Add a regression test when it protects meaningful behavior. If you claim a
   test fails before the fix, confirm that it compiles and fails for the
   intended behavioral reason.
3. Have subagents implement the fix and prove it against the reproduction and
   the affected checks.

## Keep plans and evidence out of the repository

Plans, work-in-progress notes and verification evidence do not belong in
commits. Evidence goes on the pull request.
