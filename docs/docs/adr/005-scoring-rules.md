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

We want scoring to be resolved from an ordered list of rules. Rules can match
on activity, unit, language, and tags, and are either non-stackable base rules
or stackable modifiers.

For example:

```text
base: activity=reading, unit=characters, language=jpn -> 0.0025
base: activity=reading, unit=characters -> 0.00083333
base: activity=reading, unit=page, language=jpn, tag=two_column -> 1.6
base: activity=reading, unit=page -> 1
base: activity=listening -> 0.4
modifier: activity=listening, tag=dense -> 1.5
```

Priority orders rules from most specific to least specific. The first matching
base rule supplies the base rate, so Japanese characters use `0.0025` directly
rather than stacking a language multiplier on the broad character rate. Every
matching stackable modifier is then applied; for example, dense listening
scores at `0.4 * 1.5`. A modifier cannot produce a score without a matching
base rule.

The set of activities is fixed and is not expected to change. Units still exist, but they are stable identifiers rather than containers for scoring modifiers. Contest admins may configure contest-specific scoring rules, and we also expect to tweak the platform default scoring rules over time.

## Decision

Activities will be owned by code as a fixed set of domain values.

Units will also be owned by code as stable identifiers, grouped by activity. Unit metadata should describe what can be logged, not how much it scores.

Scoring rules will be stored in the database. The platform default rules will be represented as a platform-owned rule set, using the same tables and evaluation path as contest-specific rule sets. Contest-specific rule sets can override or replace the platform defaults depending on the contest configuration.

The scoring engine will resolve a rule set, select the first matching
non-stackable base rule in priority order, and then multiply the rates of every
matching stackable modifier.

Every populated matcher must match. Rules may match the code-owned activity,
stable unit key, language code, and one normalized tag. Each rule explicitly
selects either the submitted `amount` or `duration_minutes` as its score
source. When a log contains both amount and duration, amount is authoritative
for scoring and duration remains tracking metadata.

If no base rule matches, the score is zero and the submission remains valid.
Modifiers never score without a matching base rule.

Rule sets are versioned. Drafts may be assembled and validated, but publication
makes a version immutable. Activating a new platform or contest version affects
only future score resolutions; it never recalculates stored scores.

An overriding contest rule set evaluates its rules first and falls back to a
specific published platform version pinned when the contest version is
created. A replacing contest rule set has no fallback, so uncovered inputs
score zero for that contest.

Scores should be snapshotted with the log submission that uses them. If contest scoring can differ from platform scoring, the contest score should be stored with the contest log entry rather than only on the base log.

## Consequences

Using code-owned activities and units keeps the core vocabulary stable and prevents accidental runtime changes to domain invariants.

Using database-backed scoring rules lets us tune platform defaults without redeploying application code, and lets platform defaults and contest-specific rules share one implementation path.

Snapshotting scores prevents historical leaderboards from changing silently when rules are edited. Recalculation, if needed, should be an explicit operation.

Versioning and publication add an explicit management workflow, but make
provenance durable: a score can retain the exact rule-set version, ordered rule
IDs, applied rates, and score source that produced it.

Default tags are not part of scoring metadata by default. Tags remain free text. If a tag affects scoring, it does so because a scoring rule matches it.
