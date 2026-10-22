---
title: Account-app selection
description: How each identity operation chooses tenant-local accounts or v1 Kratos from the identity-v2 flag, and when selection is unavailable.
---

# Account-app selection

Read this before changing how identity chooses its provider, or how it behaves
when the `identity-v2` decision is unavailable.

Every identity operation selects its provider from one evaluation of the
tenant's `identity-v2` flag: registration, login, logout, session lookup, the
Oathkeeper callback, recovery, settings and every `Reader` lookup.

| Flag evaluation | Selection |
| --- | --- |
| `true` from the current or last fetched snapshot | v2 tenant-local accounts |
| `false` from the current or last fetched snapshot | v1 Kratos |
| Flag absent from the fetched canonical `tadoku/prod` snapshot the client evaluated | The flag's safe default, `true` (v2) |
| Flag absent from a test tenant's fetched snapshot | Unavailable |
| No fetched snapshot, Flipt disabled, provider error or invalid response | Unavailable |
| Request canceled or past its deadline | The request's context error |

Only the canonical tenant may use the safe default, because production Flipt
need not declare the flag. A missing canonical flag therefore selects
tenant-local accounts and never calls Kratos. Tenant provisioning always creates
`identity-v2`, so a test tenant without it has lost its namespace; rerun the
tenant task. The tenant comes from the verified request context, never from a
host or cookie name.

A running client keeps evaluating its last fetched snapshot through Flipt
outages, namespace deletion and fetched snapshots that fail validation, such as
one for another namespace or without a digest. A client without one, such as after an API
restart, a startup timeout or a new or evicted test-tenant client, cannot
decide until it starts from a fetched snapshot. A snapshot that has been
received but not yet applied, that the evaluation engine rejected, or that a
client fetched only after its initial fetch failed, does not count.

An undecided selection never falls back to the other provider. It makes no
Kratos call, revokes nothing and sets no cookie. Identity routes return `503`
with `{"message":"identity service unavailable"}`, `504` when the request
deadline expired and `499` when the client canceled. The callback returns the
same statuses, so Oathkeeper denies the request instead of admitting a guest or
a Kratos identity. Logout reports the failure instead of `204`, and the native
session stays valid until a retry succeeds. Profile, moderation and ban-target
reads fail the same way.
