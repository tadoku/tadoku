---
sidebar_position: 5
title: "005 - Scoring Rules"
---

# [005] Scoring rules and metadata ownership

* Status: accepted
* Author: @antonve
* Date: 2026-06-21

## Context

Log scoring currently uses a modifier attached to the selected log unit. This makes scoring look like a property of the unit, even though future scoring needs to be more flexible.

We want scoring to be resolved from an ordered list of rules. Earlier rules are more specific, and the first matching rule is used. Rules can match on activity, unit, language, and tags.

For example:

```text
activity=reading, unit=characters, language=jpn -> 0.0025
activity=reading, unit=characters -> 0.00083333
activity=reading, unit=page, language=jpn, tag=two_column -> 1.6
activity=reading, unit=page -> 1
activity=listening, tag=dense -> 0.7
activity=listening -> 0.5
```

The set of activities is fixed and is not expected to change. Units still exist, but they are stable identifiers rather than containers for scoring modifiers. Contest admins may configure contest-specific scoring rules, and we also expect to tweak the platform default scoring rules over time.

## Decision

Activities will be owned by code as a fixed set of domain values.

Units will also be owned by code as stable identifiers, grouped by activity. Unit metadata should describe what can be logged, not how much it scores.

Scoring rules will be stored in the database. The platform default rules will be represented as a platform-owned rule set, using the same tables and evaluation path as contest-specific rule sets. Contest-specific rule sets can override or replace the platform defaults depending on the contest configuration.

The scoring engine will resolve a rule set, evaluate rules in priority order, and use the first matching rule.

Scores should be snapshotted with the log submission that uses them. If contest scoring can differ from platform scoring, the contest score should be stored with the contest log entry rather than only on the base log.

## Consequences

Using code-owned activities and units keeps the core vocabulary stable and prevents accidental runtime changes to domain invariants.

Using database-backed scoring rules lets us tune platform defaults without redeploying application code, and lets platform defaults and contest-specific rules share one implementation path.

Snapshotting scores prevents historical leaderboards from changing silently when rules are edited. Recalculation, if needed, should be an explicit operation.

Default tags are not part of scoring metadata by default. Tags remain free text. If a tag affects scoring, it does so because a scoring rule matches it.
