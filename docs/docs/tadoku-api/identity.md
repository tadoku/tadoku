---
title: Identity accounts and sessions
description: Tenant-owned accounts, password/session security, recovery and settings, account reads and remaining account contracts.
---

# Identity accounts and sessions

Read this before changing account or session behavior, implementing recovery or
settings, or moving a caller to the identity entry point.

Registration, login, logout, session lookup, recovery, profile/password settings
and the Oathkeeper session callback are implemented. Every operation, including
for guests, uses tenant-local accounts. Profile and moderation callers use
`identity.Reader`, constructed by `identity.NewReader`.

Account lookup and search are implemented by the separate `ReadService` used by
profile and moderation. The worker deletes accounts; see
[Account deletion](./account-deletion.md).
[Retained email](./email.md) provides recovery messages and old-address notices
through transactional writes and queue publication.

## HTTP boundaries

Public identity paths are relative to `/api/internal` and enter the JWT and ban
pipeline, including guest requests. Missing JWT credentials return `400`, invalid
JWTs `401`, and another tenant on a scoped deployment `421`.
`GET /identity/session` returns `account`: the signed-in account, or `null` for
a guest. It also returns `recent_login`: `true` while the session's
authentication time lies within the last hour, the window that email and
password changes require (see [Recovery and settings](#recovery-and-settings)),
and `false` for an older or future authentication time and for a guest. One
rule decides both, so a client can send the user to log in again before
a change would return `stale_session`. Registration and login return the same
session shape, with `recent_login` `true`.

Identity mutations require JSON and an HTTPS `Origin` on the Tadoku app
allowlist. App hosts include the main, account, admin and Webv3 `next.` hosts,
their preview branches and development equivalents. The former authv2 hosts
stay accepted while `authv2.tadoku.app` redirects to `account.tadoku.app`.
Credentials, explicit ports and lookalike suffixes are rejected. Return URLs use the same host policy and may
retain their path, query and fragment. Request bodies keep the shared byte cap
and generated decoder. Logout takes no body and never reads one, but still
requires the JSON `Content-Type`; whatever body arrives, its `Origin` check
comes first and its answer is never cached.

`GET /identity/internal/v1/session` requires the callback bearer
`API_OATHKEEPER_AUTHZ_TOKEN`, rather than a user JWT. It resolves canonical
`tadoku/prod` when `X-Dev-Branch` is absent, or `tadoku/<route>` when present,
checks the deployment scope, then delegates to the same identity service.
Malformed or unserved branch selections return `421`. An unrelated, expired,
revoked or invalid session returns `200` with `identity.id` equal to `guest`.
Authenticated responses carry the identity ID, its `email` and `display_name`
traits, `authenticated_at` and `expires_at`: the shape Oathkeeper's
cookie-session authenticator reads.
The caller must forward the complete Cookie header and the trusted branch
selection; neither a frontend host nor another cookie can change the tenant.

The public gateway rules must expose these paths before browser callers can
use them, and Oathkeeper must point its cookie-session authenticator at the
callback before session cookies authenticate other application routes.
Registering the API routes alone does not change the gateway configuration.

## Accounts and passwords

`Email`, `DisplayName` and `Password` have private representations and parsing
constructors. Emails are normalized to lowercase with surrounding whitespace
removed. Responses carry an account's stored address instead: the normalized
address, or an imported address as written.
When Go's `net/mail` reads an address from it, the API writes that address
part, so a quoted local part arrives unquoted and surrounding ASCII spaces and
tabs are dropped, while letter case, other characters and length are kept. An
address `net/mail` cannot read, such as a stored address with a leading
vertical tab, is written unchanged. The contract therefore types the account
`email` in identity responses and the session callback as a plain string
containing `@`, not `format: email`.
Display names contain 2–32 Unicode characters and at least one visible
character. Passwords contain at least 8 characters and at most 72 UTF-8 bytes,
the bcrypt limit. Compared in lowercase with the email, a password is rejected
when their Levenshtein distance is below 5 or their longest common substring is
longer than half the password length. Minimum length and similarity measure
Unicode characters. Password formatting is redacted.

Registration validates input, passes admission, hashes the password and
generates the initial session before acquiring a database connection. One SQL
statement creates the user, account and session atomically. User timestamps are
interpreted explicitly as UTC when stored on the account, independently of the
database session time zone. The tenant-prefixed email uniqueness constraint also
resolves concurrent registration and differently cased duplicates. A conflict
returns `409` with `email_taken`, preserving the existing enumeration behavior.
Password hashes use bcrypt cost 12 and stay inside the identity feature. Login
verifies stored hashes without applying a new registration policy to an imported
password, and returns the same `wrong_password` answer for unknown, inactive,
deleted or incorrect accounts. Unknown emails perform a cost-12 dummy
comparison. No breached-password provider is called.

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
ingress boundary. Untrusted peers cannot evade the IP budget
with forged forwarding headers. With no trusted proxy configured, clients behind
the same proxy share its peer-IP throttle bucket. This fallback prevents header spoofing
but can throttle unrelated clients together.

Registration has its own admission budget, separate from the login counters.
After input validation, and before bcrypt runs, an atomic Valkey counter per
tenant and client IP admits ten registrations per fixed
900-second window, which starts with the first valid attempt and is not extended
by later ones. It uses the same client IP and trusted-proxy rule as login, with
a tenant-prefixed, SHA-256-derived key. Duplicate and new emails both consume
the budget, so unique addresses cannot bypass it; rejected input does not.
Further attempts return `429` with `throttled` without hashing. Unlike login,
registration fails closed: every admission failure returns before hashing, with
the status the shared error mapping gives it. A Valkey error or a missing tenant
returns `503`, the admission timeout returns `504`, and a request the client
cancelled returns `499`. The limit bounds the cost-12 hashes one address can
cause while leaving room for several people to sign up behind one shared
address. Without trusted proxy configuration, every client behind one proxy
shares the budget.

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
no-store`. Development API base and overlays set
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

Recovery has its own admission, separate from the login and registration
counters. After input validation, and before the transaction opens, one atomic
Valkey script charges two tenant-prefixed, SHA-256-derived counters, each with
a fixed 900-second window from its first valid request:

- **Client:** ten requests per trusted client IP, using the login and
  registration IP rule. Further requests return `429` with `throttled` and are
  not charged to the recipient.
- **Recipient:** three requests per normalized email, from any client. Further
  requests write nothing and return the same empty `202` as an unknown
  address, so one mailbox receives at most three recovery emails per window
  however many clients ask.

Known and unknown addresses consume both counters identically, so neither
answer reveals whether an account exists. Rejected input consumes neither.
Admission fails closed with the shared error mapping, identically for known and
unknown addresses: a Valkey error or missing tenant returns `503`, the admission
timeout `504` and a cancelled request `499`, all before any token, email or job
is written. A request whose transaction later rolls back, for example on an
enqueue failure, has still used one unit of each budget; the user can retry
while the recipient budget lasts or after the window ends. The recipient budget
leaves room for two retries after a lost or delayed email within the one-hour
token lifetime. The counters use the normalized email only as a hashed key;
mail goes to the account's address as described next.

Recovery email and old-address notices go to the account's stored address
without the spaces and tabs around it. RFC 5322 reads those as folding
whitespace outside the mailbox, and identity responses drop them too. Letter
case and every other character are kept, including Unicode spaces, and the
address is never lowercased, trimmed further or truncated. The email writer then
applies its unchanged recipient rules. An address that is not a bare mailbox of
at most 320 bytes after that trim, such as a stored address with a leading
vertical tab, form feed or line break or with a long Unicode-space prefix, gets
no recovery email or notice. Recovery requests and email changes for such an
account return `400` with `invalid email recipient` and write nothing, so its
tokens, sessions and stored values stay as they were. That answer differs from
an unknown address's `202`. Login is unaffected because it uses the normalized
email.

The link uses the configured cookie domain and parsed tenant's main API host,
never the incoming Host. Its path is `/api/internal/identity/recovery/{token}`,
with the validated destination in an optional `return_to` query. Destinations
have a 2,048-byte input ceiling and must also fit, without truncation, into
the tenant's Settings URL as an encoded `return_to` after `edit=password`. The
complete Settings URL must not exceed 2,048 bytes; the encoded recovery link
must not exceed the email parameter limit of 4,096 bytes. Both bounds are
checked before any account lookup, identically for known and unknown emails.
Account hosts must use the configured main API URL rather than assume they
serve relative API paths. The token contains 256 random bits; only its SHA-256
hash is stored in `recovery_tokens`, with a one-hour expiry. The email retains
the usable link under the email feature's existing retention and
authorized-read policy.

`GET /identity/recovery/{token}` validates the destination before consuming the
token. It always redirects to that tenant's Settings root on its account host,
`account.<domain>` or the branch's account host, with `edit=password`, opening
the Password editor, and adds a supplied `return_to` as an encoded query
parameter for continuation after the user completes Settings.
An invalid or unrepresentable continuation fails before consuming the token.
Consumption
locks the active account, atomically marks the token used, invalidates its other
recovery tokens, revokes all its existing sessions and creates a fresh session.
It returns `303`, Location and the normal secure tenant cookie. The fresh
`authenticated_at` grants the recovered user the same one-hour window to set a
password. Expired, reused, malformed and foreign-tenant tokens share
`invalid_recovery_token`; simultaneous attempts, including different tokens for
one account, allow only one success.

Profile settings accept email and/or display name. Display-name changes require
a valid session; email changes additionally require authentication within the
last hour. A submitted email whose normalized form equals the address the
account's mail goes to is unchanged: nothing about the email is written and no
recent login is needed. Any other value is a change, including one that differs
from the stored address only in letter case or Unicode spaces, because it
changes where mail goes. Email switches immediately with tenant-local
normalized uniqueness, stores the normalized form, invalidates pending recovery
tokens and queues `account.email_changed` to the old address, with `new_email`
and RFC3339 `changed_at`. The response includes the
existing session authentication time. A session older than one hour, or with an
authentication time in the future, returns `stale_session` for email/password
changes. Re-authenticate through login to get a fresh cookie.

Password settings apply the same parsing and cost-12 bcrypt policy as
registration, retain the current session, revoke other sessions and invalidate
pending recovery tokens. Session lookup, recency checks, password validation and
hashing finish before the application's transaction opens. Guest settings
requests return without opening a transaction. Inside the transaction, the
account lock is followed by another session lookup, recency check and password
validation against the current email; concurrent email changes, expiry and
revocation cannot bypass these rules. Both settings writes and recovery
consumption serialize on the account row and recheck session/token state after
acquiring the lock. Login session creation locks that same row and requires the
password hash it verified still to match, preventing an in-flight old-password
login from surviving a credential replacement.

The application owns each business transaction. Identity writes, rendered email
rows and every ID-only `email.send.v1` job commit together; a rendering or enqueue
failure rolls them all back. Only a sending-enabled worker delivers, and only
`tadoku/prod` rows; development and branch workers hold mail.

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

## Account reads

`identity.NewReader(pool)` constructs the `ReadService` that profile and authz
receive as provider-free `Reader` values. Reads join active accounts to
undeleted users under the transaction-local tenant and do not share a
profile-user snapshot across requests. Each reader page uses one SQL statement,
orders by user UUID and searches display name or email with a literal
case-insensitive substring. Cursors bind the tenant, search and last UUID. No
password or token hash is selected.

The administrator HTTP list retains numbered pages and `total_size` for
compatibility. Its adapter uses the concrete reader's `ListAllAccounts`
operation, which walks the tenant-bound keyset pages to count matches and
select the requested numbered page, then resolves roles only for that page.
This can read multiple keyset pages for a large result set; the public response
shape is unchanged.
