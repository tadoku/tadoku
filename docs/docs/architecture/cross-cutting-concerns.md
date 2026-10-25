---
title: Cross-cutting concerns
description: How Tadoku API divides a concern that many features use into core logic, state, vocabulary and enforcement, where each part lives, and how to add a feature's entries or a new concern.
---

# Cross-cutting concerns

Read this when a feature uses a concern that many features share, or when you
design a new one.

A cross-cutting concern is behavior that many features need and none should
build again. It has four parts, each with one home:

| Part | What it is | Home |
| --- | --- | --- |
| Core logic | Generic logic that knows no feature names. It declares a small interface, a port, for the state it needs. | `services/tadoku-api/internal/<concern>` |
| State | A table owned by one feature's repository, or an external system behind an adapter | `services/tadoku-api/features/<owner>` or `services/tadoku-api/infra/<provider>` |
| Vocabulary | The concrete entries, such as which permissions exist | The feature they describe, or `services/tadoku-api/domain/<concept>` when one feature produces them and another consumes them |
| Enforcement | When to check, write, enqueue or record | `services/tadoku-api/app/` |

A composition root constructs the state owner, builds the core logic over it
and hands both to `app/`. Using an existing concern needs only vocabulary in the
feature and enforcement in `app/`: no core logic, no storage and no import
exception.

## Rules

- **Queries on a table live only in the repository of the feature that owns
  it,** or in an owner-only application such as `app/tenantlifecycle`.
  `internal/` and `infra/` hold no table queries.
- **Do not move code into `infra/` or `internal/` to avoid a visibility
  rule.** A concern that seems to need an import exception is a design question
  for the pull request; see
  [Import boundaries](../tadoku-api/import-boundaries.md).
- **Core logic provides the complete set of building blocks for its domain,**
  so features compose them instead of extending it. Its options and
  configuration still follow current needs.
- **Core logic imports only `internal/` packages and libraries.** It declares
  its port, the state owner implements it, and a composition root wires them. A
  port with one production implementation is allowed at this boundary; an
  interface added only to inject mocks is not. See
  [Dependencies](../tadoku-api/conventions.md#dependencies).
- **Only the state owner may wrap the concern's runtime for its own
  vocabulary.** Other features hold none of the runtime; they declare
  vocabulary and return values.
- **`app/` enforces inside the transaction of the write it follows.** Features
  return values for it to act on, such as jobs to enqueue.
- **A feature receives facts owned by another feature from `app/`,** as a value
  or a lookup, instead of importing that feature.

Feature-flag vocabulary is the exception: it is declared centrally, as
[Adding a flag](./feature-flags.md#adding-a-flag) describes.

## Add a feature's entries

1. Declare the entries in the feature, or in `domain/<concept>` when another
   feature consumes them.
2. Export typed constructors for them, so call sites never spell a name.
3. Enforce them in `app/`: checks in the operation, writes inside its
   transaction.
4. Register them where the concern's page says.
5. Leave imports and visibility unchanged.

## Design a new concern

Give each part its home from the table. When the only logic is the state's own
protocol, there is no core logic package. A new `internal/` package gets the
narrowest visibility and an entry in the `internal` policy in `.depolicy.yaml`;
see [Adding a package](../tadoku-api/import-boundaries.md#adding-a-package). Any
other import or visibility change is a design question for the pull request.
