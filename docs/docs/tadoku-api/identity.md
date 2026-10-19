---
title: Identity accounts and sessions
description: Tenant-owned accounts, password/session security, recovery and settings, v1 Kratos compatibility and remaining account contracts.
---

# Identity accounts and sessions

Read this before changing account or session behavior, implementing recovery or
settings, or moving a caller to the identity entry point.

Registration, login, logout, session lookup, recovery, profile/password settings
and the Oathkeeper session callback are implemented. A tenant's `identity-v2` flag selects local accounts or Kratos;
its safe default is `false`, including for guests. Existing profile/authz callers
continue to use `identity.LegacyReader`.

Account search/deletion remain skeletons returning `501`. V1 tenants continue
to use the existing Kratos account frontend for registration, recovery and
settings. [Retained email](./email.md) provides the recovery messages and
old-address notices through transactional writes and queue publication.

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

## Recovery and settings

`POST /identity/recovery` accepts a normalized email and an allowlisted
`return_to`. Existing active accounts receive `account.recovery`; unknown,
inactive, deleted and deletion-locked accounts receive the same empty `202` without mail.
The application pads responses to a 300 ms minimum using real elapsed time,
after the transaction releases its connection. This covers ordinary query,
rendering and enqueue differences; dependency failures still return errors and
work exceeding that minimum is not hidden.

The link uses the configured cookie domain and parsed tenant's main API host,
never the incoming Host. Its path is `/api/internal/identity/recovery/{token}`,
with the validated destination in an optional `return_to` query. Account and
authv2 hosts must use the configured main API URL rather than assume they serve
relative API paths. The token contains 256 random bits; only its SHA-256 hash is
stored in `recovery_tokens`, with a one-hour expiry. The email retains the usable
link under the email feature's existing retention and authorized-read policy.

`GET /identity/recovery/{token}` validates the destination before consuming the
token. Missing destinations use that tenant's authv2 Settings root. Consumption
locks the active account, atomically marks the token used, invalidates its other
recovery tokens, revokes all its existing sessions and creates a fresh session.
It returns `303`, Location and the normal secure tenant cookie. The fresh
`authenticated_at` grants the recovered user the same one-hour window to set a
password. Expired, reused, malformed and foreign-tenant tokens share
`invalid_recovery_token`; simultaneous attempts, including different tokens for
one account, allow only one success.

Profile settings accept email and/or display name. Display-name changes require
a valid session; email changes additionally require authentication within the
last hour. Email switches immediately with tenant-local normalized uniqueness,
invalidates pending recovery tokens and queues `account.email_changed` to the
old address, with `new_email` and RFC3339 `changed_at`. The response includes the
existing session authentication time. A session older than one hour, or with an
authentication time in the future, returns `stale_session` for email/password
changes. Re-authenticate through login to get a fresh cookie.

Password settings apply the same parsing and cost-12 bcrypt policy as
registration, retain the current session, revoke other sessions and invalidate
pending recovery tokens. Both settings writes and recovery consumption
serialize on the account row and recheck session/token state after acquiring
the lock. Login session creation locks that same row and requires the password
hash it verified still to match, preventing an in-flight old-password login
from surviving a credential replacement.

The application owns each business transaction. Identity writes, rendered email
rows and every ID-only `email.send.v1` job commit together; a rendering or enqueue
failure rolls them all back. Recovery and password changes do not mutate the
shared v1 provider. Development workers hold mail; SMTP activation remains a
separate deployment decision.

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

The temporary `identity.NewLegacyReader` retains read-only Kratos lookup,
existence and cursor pagination for existing callers. Its implementation and
provider tests live in `services/tadoku-api/features/identity/kratos/`.
