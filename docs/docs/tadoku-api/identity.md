---
title: Identity accounts and sessions
description: Tenant-owned accounts, password and session security, v1 Kratos session compatibility, and the remaining identity, email and tenant access-token contracts.
---

# Identity accounts and sessions

Read this before changing account or session behavior, implementing recovery or
settings, or moving a caller to the identity entry point.

For the standalone Kratos account importer, its disposable CLI verification and
the operational holds on shared imports, read [Import Kratos accounts](./identity-import.md).

Registration, login, logout, session lookup and the Oathkeeper session callback
are implemented. A tenant's `identity-v2` flag selects local accounts or Kratos;
its safe default is `false`, including for guests. Profile and moderation callers use
`identity.Reader`, constructed by `identity.NewReader`. It selects tenant-local
reads or the v1 adapter from the tenant flag; callers never select the provider.

Account lookup and search are implemented by the separate `ReadService` used by
profile and moderation; the `Accounts` service's embedded Reader methods remain
`501` stubs, as do recovery, settings and account deletion.
Recovery requests already reject an unsafe `return_to`. [Retained email](./email.md) implements transactional templates,
delivery, audited reads and retention; identity recovery/settings do not yet
produce those messages.

## HTTP boundaries

Public identity paths are relative to `/api/internal` and enter the JWT and ban
pipeline, including guest requests. Missing JWT credentials return `400`, invalid
JWTs `401`, and another tenant on a scoped deployment `421`. Public registration
and login return `501` for v1 tenants; their account app remains the existing
Kratos frontend. `GET /identity/session` works for both implementations and
returns `account_app` plus the account or `null`.

Identity mutations require JSON and an HTTPS `Origin` on the Tadoku app
allowlist. App hosts include the main, account, admin and authv2 hosts, their
preview branches and development equivalents. Credentials, explicit ports and
lookalike suffixes are rejected. Return URLs use the same host policy and may
retain their path, query and fragment. Request bodies keep the shared byte cap
and generated decoder.

`GET /identity/internal/v1/session` requires the callback bearer
`API_OATHKEEPER_AUTHZ_TOKEN`, rather than a user JWT. It resolves canonical
`tadoku/prod` when `X-Dev-Branch` is absent, or `tadoku/<route>` when present,
checks the deployment scope, then delegates to the same identity service.
Malformed or unserved branch selections return `421`. An unrelated, expired,
revoked or invalid session returns `200` with `identity.id` equal to `guest`.
Authenticated responses preserve Kratos's whoami identity/traits/timestamps shape.
The caller must forward the complete Cookie header and the trusted branch
selection; neither a frontend host nor another cookie can change the tenant.

The public gateway rules must expose these paths before browser callers can
use them, and Oathkeeper must point its cookie-session authenticator at the
callback before v2 cookies authenticate other application routes. Registering
the API routes alone does not change the gateway configuration.

## Accounts and passwords

`Email`, `DisplayName` and `Password` have private representations and parsing
constructors. Emails are normalized to lowercase with surrounding whitespace
removed. Display names contain 2–32 Unicode characters and at least one visible
character. Passwords contain at least 8 characters, use at most 72 UTF-8 bytes,
and satisfy the Kratos v26.2.0 similarity rule: compare lowercase values, reject
Levenshtein distance below 5 or a longest common substring exceeding half the
password length. Minimum length and similarity measure Unicode characters;
Kratos uses byte lengths for its minimum and substring ratio, so multibyte
password acceptance can differ. The explicit 72-byte bcrypt limit remains the
same. Password formatting is redacted.

Registration evaluates the tenant flag, validates input, hashes the password
and generates the initial session before acquiring a database connection. One
SQL statement creates the user, account and session atomically. User timestamps
are interpreted explicitly as UTC when stored on the account, independently of
the database session time zone. The tenant-prefixed email uniqueness constraint
also resolves concurrent registration and differently cased duplicates. A
conflict returns `409` with `email_taken`, preserving the existing enumeration
behavior. Password hashes use bcrypt cost 12 and stay inside the identity
feature. Login verifies stored hashes without applying a new registration
policy to an imported password, and returns the same `wrong_password` answer for
unknown, inactive, deleted or incorrect accounts. Unknown emails perform a
cost-12 dummy comparison. No breached-password provider is called.

Login attempts use atomic Valkey counters per normalized email and client
IP, with tenant-prefixed, SHA-256-derived keys. A 15-minute window admits ten
attempts per email and fifty per IP; further attempts return `429` with
`throttled`. Counters include successful attempts. Provider errors log a warning
without the email, password or token and allow the attempt. The HTTP boundary
uses the connection peer unless it belongs to an explicitly configured
`API_IDENTITY_TRUSTED_PROXY_CIDRS` prefix. Only those peers may supply a single,
valid `X-Real-IP`; the trusted edge must overwrite that header. Empty proxy
configuration trusts no forwarded address. The development base and overlays
configure the cluster pod prefix; production must configure its own trusted
ingress boundary before activation. Untrusted peers cannot evade the IP budget
with forged forwarding headers. With no trusted proxy configured, clients behind
the same proxy share its peer-IP throttle bucket. This fallback prevents header spoofing
but can throttle unrelated clients together.

Registration does not use the login counters. It still performs cost-12 bcrypt
work per valid submission; registration abuse controls remain a security-review
and caller-flow concern before activation.

## Sessions and cookies

Each registration or login generates a fresh 256-bit secret. Only its SHA-256
hash is persisted, together with its tenant, account, authentication time and
expiry. Lookups join active accounts and available users inside the request's
RLS scope. Logout revokes only the selected session, remains idempotent, and
clears that tenant's cookie.

| Tenant | Cookie | Lifetime |
| --- | --- | --- |
| `tadoku/prod` | `tadoku_session` | 720 hours |
| Other parsed tenants | `tadoku_session_` plus the key with `/` replaced by `_` | 7 days |

Cookies use path `/`, `HttpOnly`, `Secure` and `SameSite=Lax`, with parent domain
`tadoku.app` or `tadoku.dev.lab`. Both production and branch cookies may be
present: the service selects the exact name for its context tenant. Renaming a
foreign token cannot cross the database tenant boundary. Duplicate selected
cookie names fail closed. Session responses use `Cache-Control: private,
no-store`.

The v1 adapter selects only `ory_kratos_session`, checks it against the Kratos
public API and maps its real account timestamps and traits. Logout deactivates
that specific session through the admin API only for the canonical tenant.
Test tenants clear the browser cookie without mutating shared Kratos state.
`API_KRATOS_PUBLIC_URL` must be configured before switching callers to v1 session
lookup; an unset URL fails closed for a present Kratos cookie. Development API
base and overlays supply this URL and
`API_IDENTITY_COOKIE_DOMAIN=tadoku.dev.lab`, including frontend-only branches
served by the base API.

## Contracts for the remaining operations

`services/tadoku-api/features/identity/` owns `Reader` and `Accounts`. Its
`CookieHeader` arguments contain the complete incoming Cookie header, including
names, rather than a token or bearer. `RecoveryToken` contains the raw URL token.
Recovery/profile results return `domain/mail.Message` values; application code
coordinates identity writes, email writes and queue jobs in one transaction.
Features do not import sibling features or expose credential hashes.

`features/email` owns template rendering, persistence, sending, retention and
reads. Tenant access-token creation, listing and revocation are implemented; see
[Email access tokens](./tenant-lifecycle.md#email-access-tokens).
`domain/mail` declares `account.recovery` and `account.email_changed`;
`domain/jobs.SendEmailV1` carries only the email ID. Email-token routes enter
`Application.AuthenticateEmailReader`, which verifies an unrevoked, unexpired
`email:read` token in its own tenant; see
[Email access tokens](../architecture/authorization.md#email-access-tokens).
The email gateway rule forwards `tdk_at_` bearers while preserving administrator
cookie authentication. A token-shaped bearer on every other public route remains
an invalid JWT.

## V1 adapter

`identity.NewLegacyReader` constructs the unchanged read-only Kratos adapter
behind `LegacyReader`. Its raw implementation and provider tests live in
`services/tadoku-api/features/identity/kratos/`. Only identity ownership and the owned Kratos fixture consume this temporary
interface. Profile and authz receive provider-free `Reader` values. Reads do not mutate Kratos or share a profile-user snapshot across requests.

V1 reads retain the user-schema filter, trait mapping, provider order and fuzzy
search ranking. Listing reads the provider's complete cursor sequence per call,
deduplicates IDs and rejects repeated continuation tokens; no stale snapshot is
served after an outage. V2 reads join active accounts to undeleted users under
the transaction-local tenant. Each reader page uses one SQL statement, orders by
user UUID and searches display name or email with a literal case-insensitive
substring. Cursors bind the tenant, search and last UUID. No password or token
hash is selected.

The administrator HTTP list retains numbered pages and `total_size` for
compatibility. Its adapter walks reader pages to count matches and select the
requested numbered page, then resolves roles only for that page. This can read
multiple keyset pages for a large result set; the public response shape is
unchanged.

The numbered admin compatibility list uses the concrete reader's
`ListAllAccounts` operation. A v1 request performs one complete Kratos cursor
walk, maps and deduplicates identities in provider order, then applies the legacy
fuzzy search and numbered page. Its snapshot exists only for that request.
V2 traverses tenant-bound SQL keyset pages; each page is one statement joining
active accounts to undeleted users. Frozen `Reader` and `Accounts` signatures
remain unchanged.
