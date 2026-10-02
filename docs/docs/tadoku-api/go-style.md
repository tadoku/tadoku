---
title: Go style
description: Layout rules for handwritten Go in the repository, covering line length, wrapping, blank-line grouping, composite literals, test structure and SQL embedded in Go.
sidebar_position: 2.5
---

# Go style

Read this before you write or change any handwritten Go, including tests.
Comment rules are in [Conventions](./conventions.md#readability).

`gofmt` is the floor, not the bar. It passes code that is hard to read, so
review applies these rules on top of it. They apply to new and changed code;
do not reformat untouched code in an unrelated change.

## Line length and wrapping

- Keep lines under about 120 columns. Wrap anything longer rather than relying
  on horizontal scrolling.
- A signature that does not fit puts one parameter per line, with the closing
  parenthesis and results on their own line:

  ```go
  func NewJWTAuthentication(
  	lifetime context.Context,
  	jwksURL string,
  	timeout, maxTokenAge time.Duration,
  	issuer string,
  	deployment tenant.Deployment,
  	logger *slog.Logger,
  ) (func(stdhttp.Handler) stdhttp.Handler, error) {
  ```

- A call that does not fit puts one argument per line with a trailing comma.
  A long format string goes on its own line, followed by its arguments.
- A long boolean condition breaks after `&&` or `||`, one clause per line.
- Do not pack a call, a scan and a compound check into one
  `if x := ...; cond` line. Assign first, then check:

  ```go
  err = executor.QueryRow(ctx, currentTenantSQL).Scan(&current)
  if err != nil || current != want {
  ```

- Split a long string literal with `+` only when the joined text stays
  identical.

## Grouping with blank lines

- A function body is a sequence of cohesive groups of roughly 3 to 12 lines,
  each doing one thing, separated by single blank lines. A run of more than
  about 15 lines without a break needs a reason, such as one statement or one
  data table.
- Typical groups: validation, setup, the main work, cleanup and result
  mapping. In handlers, separate authorization, input extraction, persistence
  and response mapping.
- Keep an `if err != nil` check attached to the statement that produced the
  error, and a `defer` attached to the acquisition it releases.
- Do not add a blank line after every statement, at the start or end of a
  block, or twice in a row.

## Composite literals and data

- Put struct fields and composite-literal entries on separate lines when the
  literal does not fit on one short line. Do not pack several key-value pairs
  per line.
- Test-case tables put one case per line, or one field per line for larger
  cases. Keep the shape of entries consistent within a table.
- A SQL string or other literal repeated in a function or file becomes a named
  constant at the narrowest scope that covers its uses.

## Tests

- `t.Helper()` and `t.Parallel()` are the first statements in a test or helper.
  Derived contexts such as `tenantCtx` follow them.
- Group a test into arrange, act and assert, with a blank line between each.
  A test with several scenarios separates each scenario as its own group.
- Test structure and fixture rules are in [Testing](./testing.md).

## SQL in Go

- A statement that does not fit on one line is a multi-line raw string,
  one clause per line, indented one tab inside the call:

  ```go
  err = db.Pool.QueryRow(ctx, `select count(*), count(*) filter (where actual <> expected)
  	from tenant_transport_rows`).Scan(&stored, &misfiled)
  ```

- A fixture `insert` with several rows puts one row per line. When rows must
  wrap, wrap every row at the same columns so the rows still read as a table.
- Arguments to a multi-line statement go on their own line after the string.
- SQL keywords follow [SQL style](./database.md#sql-style).
